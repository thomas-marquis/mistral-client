package mlflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	logger = log.New(os.Stdout, "mlflow-client: ", log.LstdFlags|log.Lshortfile)
)

// Version represents either a version (e.g. "1", "2", etc.) or an alias when prefixed by "@" (e.g. "@production").
type Version string

func (v Version) String() string {
	return string(v)
}

func (v Version) IsAlias() bool {
	return strings.HasPrefix(v.String(), "@")
}

func (v Version) TrimAlias() string {
	return strings.TrimPrefix(v.String(), "@")
}

const (
	VersionLatest Version = "latest"
)

type PromptRegistry interface {
	// Get retrieves a prompt by name and version from the prompt registry.
	// If the version is prefixed by "@", it will be treated as an alias. E.g. version = "@production" => alias = "production".
	Get(ctx context.Context, name string, version Version, opts ...PromptOption) (Prompt, error)
}

type promptRegistryImpl struct {
	mlflowUrl  string
	verbose    bool
	httpClient *http.Client
}

type PromptRegistryOption func(*promptRegistryImpl)

func WithVerbose(verbose bool) PromptRegistryOption {
	return func(r *promptRegistryImpl) {
		r.verbose = verbose
	}
}

func WithHttpClient(client *http.Client) PromptRegistryOption {
	return func(r *promptRegistryImpl) {
		r.httpClient = client
	}
}

func NewPromptRegistry(mlflowUrl string, opts ...PromptRegistryOption) (PromptRegistry, error) {
	r := &promptRegistryImpl{
		mlflowUrl:  strings.TrimSuffix(mlflowUrl, "/"),
		httpClient: http.DefaultClient,
	}

	res, err := r.httpClient.Get(fmt.Sprintf("%s/health", r.mlflowUrl))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mlflow server: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mlflow server is not healthy: %d", res.StatusCode)
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func (r *promptRegistryImpl) Get(_ context.Context, name string, version Version, opts ...PromptOption) (Prompt, error) {
	// TODO: use context to cancel request
	// TODO: handle retries and timeout
	if version == "" {
		version = VersionLatest
	}

	qp := url.Values{}
	qp.Add("name", name)

	var epUrl string
	if version == VersionLatest || version.IsAlias() {
		qp.Add("alias", version.TrimAlias())
		epUrl = fmt.Sprintf("%s/api/2.0/mlflow/registered-models/alias?%s", r.mlflowUrl, qp.Encode())
	} else {
		qp.Add("version", version.String())
		epUrl = fmt.Sprintf("%s/api/2.0/mlflow/model-versions/get?%s", r.mlflowUrl, qp.Encode())
	}

	if r.verbose {
		logger.Printf("Getting prompt %s (version: %s) from %s", name, version, epUrl)
	}

	resp, err := r.httpClient.Get(epUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get registered model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		apiErr.StatusCode = resp.StatusCode
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("mlflow api returned status %d", resp.StatusCode)
		}
		return nil, &apiErr
	}

	var wrapper struct {
		ModelVersion modelVersion `json:"model_version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var content string
	var promptType string
	for _, tag := range wrapper.ModelVersion.Tags {
		if tag.Key == "mlflow.prompt.text" {
			content = tag.Value
		}
		if tag.Key == "_mlflow_prompt_type" {
			promptType = tag.Value
		}
	}

	var prompt Prompt
	switch promptType {
	case "text":
		prompt = NewPromptText(name, Version(wrapper.ModelVersion.Version), content, opts...)
	case "chat":
		var messages []PromptChatMessage
		if err := json.Unmarshal([]byte(content), &messages); err != nil {
			return nil, err
		}
		prompt = NewPromptChat(name, Version(wrapper.ModelVersion.Version), messages, opts...)
	default:
		return nil, fmt.Errorf("unsupported prompt type: %s", promptType)
	}

	return prompt, nil
}
