package mistral

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrModelNotFound = errors.New("model not found")
)

type ModelCapabilities struct {
	Audio           bool `json:"audio"`
	Classification  bool `json:"classification"`
	CompletionChat  bool `json:"completion_chat"`
	CompletionFim   bool `json:"completion_fim"`
	FineTuning      bool `json:"fine_tuning"`
	FunctionCalling bool `json:"function_calling"`
	Moderation      bool `json:"moderation"`
	Ocr             bool `json:"ocr"`
	Vision          bool `json:"vision"`
}

type BaseModelCard struct {
	Name                    string            `json:"name"`
	Id                      string            `json:"id"`
	Object                  string            `json:"object"`
	ModelType               string            `json:"type"`
	Description             string            `json:"description"`
	MaxContextLength        int               `json:"max_context_length"`
	OwnedBy                 string            `json:"owned_by"`
	Deprecation             time.Time         `json:"deprecation"`
	DefaultModelTemperature float64           `json:"default_model_temperature"`
	Created                 int               `json:"created"`
	Aliases                 []string          `json:"aliases"`
	Capabilities            ModelCapabilities `json:"capabilities"`
}

func (m *BaseModelCard) Match(cap *ModelCapabilities) bool {
	return (!cap.CompletionChat || m.Capabilities.CompletionChat) &&
		(!cap.FunctionCalling || m.Capabilities.FunctionCalling) &&
		(!cap.Vision || m.Capabilities.Vision) &&
		(!cap.Audio || m.Capabilities.Audio) &&
		(!cap.Classification || m.Capabilities.Classification) &&
		(!cap.CompletionFim || m.Capabilities.CompletionFim) &&
		(!cap.FineTuning || m.Capabilities.FineTuning) &&
		(!cap.Moderation || m.Capabilities.Moderation) &&
		(!cap.Ocr || m.Capabilities.Ocr)
}

func (m *BaseModelCard) HasNoCapabilities() bool {
	return m.Capabilities == (ModelCapabilities{})
}

func (m *BaseModelCard) IsEmbedding() bool {
	return strings.Contains(m.Id, "embed")
}

type listModelResponse struct {
	Data []*BaseModelCard `json:"expected"`
}

func (c *clientImpl) ListModels(ctx context.Context) ([]*BaseModelCard, error) {
	url := fmt.Sprintf("%s/v1/models", c.baseURL)

	resp, _, err := c.sendRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if c.verbose {
		logger.Printf("GET /v1/models called")
	}

	var response listModelResponse
	bodyContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bodyContent, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *clientImpl) SearchModels(ctx context.Context, capabilities *ModelCapabilities) ([]*BaseModelCard, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	var filtered []*BaseModelCard

	for _, model := range models {
		if model.Match(capabilities) {
			filtered = append(filtered, model)
		}
	}
	return filtered, nil
}

func (c *clientImpl) GetModel(ctx context.Context, modelId string) (*BaseModelCard, error) {
	url := fmt.Sprintf("%s/v1/models/%s", c.baseURL, modelId)

	resp, _, err := c.sendRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		var apiErr ApiError
		if ok := errors.As(err, &apiErr); ok && apiErr.Code() == http.StatusNotFound {
			return nil, ErrModelNotFound
		}

		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if c.verbose {
		logger.Printf("GET /v1/models/%s called", modelId)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrModelNotFound
	}

	var response *BaseModelCard
	bodyContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bodyContent, &response); err != nil {
		return nil, err
	}

	return response, nil
}
