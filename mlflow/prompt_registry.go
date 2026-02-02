package mlflow

import (
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

type Version string

func (v Version) String() string {
	return string(v)
}

const (
	VersionLatest Version = "latest"
)

type PromptRegistry interface {
	Get(name string, version Version) (Prompt, error)
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

func NewPromptRegistry(mlflowUrl string, opts ...PromptRegistryOption) PromptRegistry {
	r := &promptRegistryImpl{
		mlflowUrl:  strings.TrimSuffix(mlflowUrl, "/"),
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *promptRegistryImpl) Get(name string, version Version) (Prompt, error) {
	if version == "" {
		version = VersionLatest
	}

	qp := url.Values{}
	qp.Add("name", name)

	var epUrl string
	if version == VersionLatest {
		qp.Add("alias", VersionLatest.String())
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
	for _, tag := range wrapper.ModelVersion.Tags {
		if tag.Key == "mlflow.prompt.text" {
			content = tag.Value
			break
		}
	}

	prompt := NewPromptText(name, Version(wrapper.ModelVersion.Version), content)

	return prompt, nil
}
