package mistral

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CompletionConfig struct {
	// MaxTokens is the maximum number of tokens to generate in the completion. The token count of your prompt plus max_tokens cannot exceed the model's context length.
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature to use, we recommend between 0.0 and 0.7.
	//
	// Higher values like 0.7 will make the output more random, while lower values like 0.2 will make it more focused and deterministic.
	// We generally recommend altering this or top_p but not both.
	// The default value varies depending on the model you are targeting.
	// Call the /models endpoint to retrieve the appropriate value.
	Temperature float64 `json:"temperature,omitempty"`

	// TopP is the nucleus sampling, where the model considers the results of the tokens with top_p probability mass.
	//
	// So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	// We generally recommend altering this or temperature but not both.
	// Default to 1.0
	TopP float64 `json:"top_p,omitempty"`

	// ResponseFormat specifies the format that the model must output.
	//
	// By default, it will use ResponseFormat.Type = ResponseFormatText.
	// Setting to ResponseFormat.Type = ResponseFormatJsonObject enables JSON mode, which guarantees the message the model generates is in JSON.
	// When using JSON mode you MUST also instruct the model to produce JSON yourself with a system or a user message.
	// Setting to ResponseFormat.Type = ResponseFormatJsonSchema enables JSON schema mode, which guarantees the message the model generates is in JSON and follows the schema you provide.
	// Prefer using the options builder WithResponseTextFormat, WithResponseJsonSchema, WithResponseJsonObjectFormat instead of setting this field directly.
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// ToolChoice controls which (if any) tool is called by the model.
	//
	//  - ToolChoiceNone means the model will not call any tool and instead generates a message.
	//  - ToolChoiceAuto means the model can pick between generating a message or calling one or more tools.
	//  - ToolChoiceAny or required means the model must call one or more tools.
	//
	// Specifying a particular tool via \{"type": "function", "function": \{"name": "my_function"\}\} forces the model to call that tool.
	// You can marshal a ToolChoice object directly into this field.
	ToolChoice ToolChoiceType `json:"tool_choice,omitempty"`

	// ParallelToolCalls defines whether to enable parallel function calling during tool use, when enabled the model can call multiple tools in parallel. Default to true when NewChatCompletionRequest is used.
	ParallelToolCalls bool `json:"parallel_tool_calls,omitempty"`

	// FrequencyPenalty penalizes the repetition of words based on their frequency in the generated text. A higher frequency penalty discourages the model from repeating words that have already appeared frequently in the output, promoting diversity and reducing repetition.
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`

	// PresencePenalty determines how much the model penalizes the repetition of words or phrases. A higher presence penalty encourages the model to use a wider variety of words and phrases, making the output more diverse and creative.
	PresencePenalty float64 `json:"presence_penalty,omitempty"`

	// N is the number of completions to return for each request, input tokens are only billed once.
	N int `json:"n,omitempty"`

	// PromptMode allows toggling between the reasoning mode and no system prompt. When set to reasoning the system prompt for reasoning models will be used. Default to "".
	PromptMode string `json:"prompt_mode,omitempty"`

	// RandomSeed is the seed to use for random sampling. If set, different calls will generate deterministic results.
	RandomSeed int `json:"random_seed,omitempty"`

	// SafePrompt defines whether to inject a safety prompt before all conversations.
	SafePrompt bool `json:"safe_prompt,omitempty"`

	// Stop generation if this token is detected. Or if one of these tokens is detected when providing an array
	Stop []string `json:"stop,omitempty"`

	// Stream defines whether to stream back partial progress.
	//
	// If set, tokens will be sent as data-only server-side events as they become available, with the stream terminated by a data: [DONE] message.
	// Otherwise, the server will hold the request open until the timeout or until completion, with the response containing the full result as JSON.
	Stream bool `json:"stream,omitempty"`
}

type ChatCompletionRequest struct {
	CompletionConfig

	// Model is the ID of the model to use. You can use the ListModels method to see all of your available models, or see https://docs.mistral.ai/getting-started/models overview for model descriptions.
	Model string `json:"model"`

	// Messages is(are) the prompt(s) to generate completions for, encoded as a list of dict with role and content.
	Messages []ChatMessage `json:"messages"`

	Tools []Tool `json:"tools,omitempty"`
}

type ChatCompletionRequestOption func(request *ChatCompletionRequest)

func NewChatCompletionRequest(model string, messages []ChatMessage, opts ...ChatCompletionRequestOption) *ChatCompletionRequest {
	r := &ChatCompletionRequest{
		CompletionConfig: CompletionConfig{
			ParallelToolCalls: true,
			Stream:            false,
		},
		Messages: messages,
		Model:    model,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

type ChatCompletionResponse struct {
	Choices []ChatCompletionChoice `json:"choices"`
	Created time.Time              `json:"created"`
	Id      string                 `json:"id"`
	Model   string                 `json:"model"`
	Object  string                 `json:"object"`
	Usage   UsageInfo              `json:"usage"`

	Latency time.Duration
}

var _ json.Unmarshaler = (*ChatCompletionResponse)(nil)

func (r *ChatCompletionResponse) UnmarshalJSON(data []byte) error {
	type Alias ChatCompletionResponse
	aux := &struct {
		*Alias
		Created int64 `json:"created"`
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Created = time.Unix(aux.Created, 0).UTC()
	return nil
}

// AssistantMessage returns the first assistant message in the response choices, or nil if there are no assistant messages.
func (r *ChatCompletionResponse) AssistantMessage() *AssistantMessage {
	if len(r.Choices) == 0 {
		logger.Printf("No choices found in response")
		return nil
	}
	msg := r.Choices[0].Message
	if msg == nil {
		logger.Printf("Message is nil")
		return nil
	}
	if msg.MessageRole != RoleAssistant {
		logger.Printf("Message is not an assistant message: %+v", msg)
		return nil
	}
	return msg
}

type ChatCompletionChoice struct {
	FinishReason FinishReason      `json:"finish_reason"`
	Index        int               `json:"index"`
	Message      *AssistantMessage `json:"message"`
}

type FinishReason string

const (
	FinishReasonLength      FinishReason = "length"
	FinishReasonStop        FinishReason = "stop"
	FinishReasonModelLength FinishReason = "model_length"
	FinishReasonError       FinishReason = "error"
	FinishReasonToolCalls   FinishReason = "tool_calls"
)

type ResponseFormatType string

const (
	ResponseFormatText       ResponseFormatType = "text"
	ResponseFormatJsonObject ResponseFormatType = "json_object"
	ResponseFormatJsonSchema ResponseFormatType = "json_schema"
)

type ResponseFormat struct {
	Type       ResponseFormatType `json:"type"`
	JsonSchema *JsonSchema        `json:"json_schema,omitempty"`
}

// WithResponseTextFormat ensures the response is formatted as text.
// This is the default format if none is specified.
func WithResponseTextFormat() ChatCompletionRequestOption {
	return func(req *ChatCompletionRequest) {
		req.ResponseFormat = &ResponseFormat{Type: ResponseFormatText, JsonSchema: nil}
	}
}

// WithResponseJsonSchema ensures the response is formatted according to the specified JSON schema.
// You MUST also instruct the model to produce JSON yourself with a system or a user message.
func WithResponseJsonSchema(schema any) ChatCompletionRequestOption {
	return func(req *ChatCompletionRequest) {
		req.ResponseFormat = &ResponseFormat{
			Type: ResponseFormatJsonSchema,
			JsonSchema: &JsonSchema{
				Name:   "responseJsonSchema",
				Schema: schema,
				Strict: true,
			},
		}
	}
}

// WithResponseJsonObjectFormat ensures the response is formatted as a JSON object.
// You MUST also instruct the model to produce JSON yourself with a system or a user message.
func WithResponseJsonObjectFormat() ChatCompletionRequestOption {
	return func(req *ChatCompletionRequest) {
		req.ResponseFormat = &ResponseFormat{Type: ResponseFormatJsonObject, JsonSchema: nil}
	}
}

// WithTools enables the model to call the specified tools.
func WithTools(tools []Tool) ChatCompletionRequestOption {
	return func(req *ChatCompletionRequest) {
		req.Tools = tools
		req.ToolChoice = ToolChoiceAuto
	}
}

// WithToolChoice controls which (if any) tool is called by the model.
func WithToolChoice(toolChoice ToolChoiceType) ChatCompletionRequestOption {
	return func(req *ChatCompletionRequest) {
		req.ToolChoice = toolChoice
	}
}

// WithStreaming enables streaming back partial progress.
func WithStreaming() ChatCompletionRequestOption {
	return func(req *ChatCompletionRequest) {
		req.Stream = true
	}
}

// ChatCompletion send a /chat/completion request to Mistral API and return the response.
// Messages is(are) the prompt(s) to generate completions for, encoded as a list of dict with role and content.
// Model is the name of the model to use (e.g. "mistral-small-latest"). You can use the ListModels method to see all of your available models, or see https://docs.mistral.ai/getting-started/models overview for model descriptions.
func (c *clientImpl) ChatCompletion(
	ctx context.Context,
	req *ChatCompletionRequest,
) (*ChatCompletionResponse, error) {
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/v1/chat/completions", c.baseURL)

	if req.Stream {
		return nil, fmt.Errorf("the method ChatCompletion does not support streaming")
	}

	jsonValue, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	response, lat, err := c.sendRequest(ctx, http.MethodPost, url, jsonValue)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() //nolint:errcheck

	if c.verbose {
		logger.Printf("POST /v1/chat/completions called")
	}

	var resp ChatCompletionResponse
	if err := unmarshallBody(response, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	resp.Latency = lat

	return &resp, nil
}

type CompletionResponseStreamChoice struct {
	FinishReason FinishReason      `json:"finish_reason"`
	Index        int               `json:"index"`
	Delta        *AssistantMessage `json:"delta"`
}

type CompletionChunk struct {
	Choices []CompletionResponseStreamChoice `json:"choices"`
	Created time.Time                        `json:"created"`
	Id      string                           `json:"id"`
	Model   string                           `json:"model"`
	Object  string                           `json:"object"`
	Usage   UsageInfo                        `json:"usage"`

	IsLastChunk  bool          `json:"-"`
	ChunkLatency time.Duration `json:"-"`
	TotalLatency time.Duration `json:"-"`
	Error        error         `json:"-"`
}

var _ json.Unmarshaler = (*CompletionChunk)(nil)

func (c *CompletionChunk) UnmarshalJSON(data []byte) error {
	type Alias CompletionChunk
	aux := &struct {
		*Alias
		Created int64 `json:"created"`
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.Created = time.Unix(aux.Created, 0).UTC()
	return nil
}

func (c *CompletionChunk) DeltaMessage() AssistantMessage {
	if len(c.Choices) == 0 {
		return AssistantMessage{}
	}
	return *c.Choices[0].Delta
}

func (c *clientImpl) ChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (<-chan *CompletionChunk, error) {
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/v1/chat/completions", c.baseURL)
	if !req.Stream {
		return nil, fmt.Errorf("the method ChatCompletionStream requires streaming")
	}

	jsonValue, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	outChan := make(chan *CompletionChunk)

	res, lat, err := c.sendRequest(ctx, http.MethodPost, url, jsonValue)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(outChan)
		defer res.Body.Close()

		buff := bufio.NewReader(res.Body)

		var t0 time.Time
		var i uint
		totLat := lat
		for {
			t0 = time.Now()
			line, err := buff.ReadBytes('\n')
			lat += time.Since(t0)
			if err != nil {
				if err == io.EOF {
					return
				}
				outChan <- &CompletionChunk{Error: fmt.Errorf("failed to read response line: %w", err)}
				return
			}

			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			jsonPart := strings.TrimSpace(
				strings.TrimPrefix(string(line), "data: "))

			if jsonPart == "[DONE]" {
				return
			}
			var chunk CompletionChunk
			if err := json.Unmarshal([]byte(jsonPart), &chunk); err != nil {
				outChan <- &CompletionChunk{
					Error: fmt.Errorf("failed to unmarshal response chunk %d '%s': %w", i, jsonPart, err)}
				return
			}
			chunk.ChunkLatency = lat
			totLat += lat
			lat = 0
			if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
				chunk.IsLastChunk = true
				chunk.TotalLatency = totLat
			}
			outChan <- &chunk
			i++
		}
	}()

	return outChan, nil

}
