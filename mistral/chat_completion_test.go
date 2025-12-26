package mistral_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/mistral-client/mistral"
)

func TestChatCompletion(t *testing.T) {
	t.Run("Should call Mistral /chat/completion endpoint", func(t *testing.T) {
		var gotReq string
		mockServer := makeMockServerWithCapture(t, "POST", "/v1/chat/completions", `
				{
					"id": "1234567",
					"created": 1764230687,
					"model": "mistral-small-latest",
					"usage": {
						"prompt_tokens": 13,
						"total_tokens": 23,
						"completion_tokens": 10
					},
					"object": "chat.completion",
					"choices": [
						{
							"index": 0,
							"finish_reason": "stop",
							"message": {
								"role": "assistant",
								"tool_calls": null,
								"content": "Hello, how can I assist you?"
							}
						}
					]
				}`, http.StatusOK, &gotReq)
		defer mockServer.Close()

		// Given
		ctx := context.TODO()
		c := mistral.New("fakeApiKey", mistral.WithBaseApiUrl(mockServer.URL))
		inputMsgs := []mistral.ChatMessage{
			mistral.NewSystemMessageFromString("You are a helpful assistant."),
			mistral.NewUserMessageFromString("Hello!"),
		}

		// When
		res, err := c.ChatCompletion(ctx, mistral.NewChatCompletionRequest("mistral-small-latest", inputMsgs))

		// Then
		assert.NoError(t, err)
		assert.Len(t, res.Choices, 1)
		assert.Equal(t, mistral.NewAssistantMessageFromString("Hello, how can I assist you?"), res.Choices[0].Message)

		// Check usage
		assert.Equal(t, 13, res.Usage.PromptTokens)
		assert.Equal(t, 23, res.Usage.TotalTokens)
		assert.Equal(t, 10, res.Usage.CompletionTokens)

		assert.Equal(t, "chat.completion", res.Object)
		assert.Equal(t, "mistral-small-latest", res.Model)

		expectedReq := `{
		  "model": "mistral-small-latest",
		  "messages": [
			{
			  "role": "system",
			  "content": "You are a helpful assistant."
			},
			{
			  "role": "user",
			  "content": "Hello!"
			}
		  ],
		  "parallel_tool_calls": true
		}`
		assert.JSONEq(t, expectedReq, gotReq)
	})

	t.Run("should call Mistral with tools, tool choice and multiple messages", func(t *testing.T) {
		var gotReq string
		mockServer := makeMockServerWithCapture(t, "POST", "/v1/chat/completions", `
				{
					"id": "12345",
					"created": 1764282082,
					"model": "mistral-small-latest",
					"usage": {
						"prompt_tokens": 142,
						"total_tokens": 158,
						"completion_tokens": 16
					},
					"object": "chat.completion",
					"choices": [
						{
							"index": 0,
							"finish_reason": "tool_calls",
							"message": {
								"role": "assistant",
								"tool_calls": [
									{
										"id": "abcde",
										"function": {
											"name": "add",
											"arguments": "{\"a\": 2, \"b\": 3}"
										},
										"index": 0
									}
								],
								"content": ""
							}
						}
					]
				}`, http.StatusOK, &gotReq)
		defer mockServer.Close()

		// Given
		ctx := context.TODO()
		c := mistral.New("fakeApiKey", mistral.WithBaseApiUrl(mockServer.URL))
		inputMsgs := []mistral.ChatMessage{
			mistral.NewSystemMessageFromString("You are a helpful assistant."),
			mistral.NewUserMessageFromString("2 + 3?"),
		}

		// When
		res, err := c.ChatCompletion(ctx,
			mistral.NewChatCompletionRequest(
				"mistral-small-latest",
				inputMsgs,
				mistral.WithTools([]mistral.Tool{
					mistral.NewTool("add", "add two numbers", map[string]any{
						"type": "object",
						"properties": map[string]any{
							"a": map[string]any{
								"type": "number",
							},
							"b": map[string]any{
								"type": "number",
							},
						},
					}),
					mistral.NewTool("getUserById", "get user by id", map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id": map[string]any{
								"type": "string",
							},
						},
					}),
				}),
				mistral.WithToolChoice(mistral.ToolChoiceAny),
			),
		)

		// Then
		assert.NoError(t, err)
		assert.Len(t, res.Choices, 1)
		assert.Equal(t, mistral.RoleAssistant, res.Choices[0].Message.Type())
		assert.Equal(t, mistral.FinishReasonToolCalls, res.Choices[0].FinishReason)
		assert.Len(t, res.Choices[0].Message.ToolCalls, 1)
		assert.Equal(t, "add", res.Choices[0].Message.ToolCalls[0].Function.Name)
		assert.Equal(t, mistral.JsonMap{"a": 2., "b": 3.}, res.Choices[0].Message.ToolCalls[0].Function.Arguments)

		expectedReq := `{
		  "model": "mistral-small-latest",
		  "messages": [
			{
			  "role": "system",
			  "content": "You are a helpful assistant."
			},
			{
			  "role": "user",
			  "content": "2 + 3?"
			}
		  ],
		  "tools": [
			{
			  "type": "function",
			  "function": {
				"name": "add",
				"description": "add two numbers",
				"parameters": {
				  "type": "object",
				  "description": "",
				  "properties": {
					"a": {"type": "number", "description": ""},
					"b": {"type": "number", "description": ""}
				  }
				}
			  }
			},
			{
			  "type": "function",
			  "function": {
				"name": "getUserById",
				"description": "get user by id",
				"parameters": {
				  "type": "object",
				  "description": "",
				  "properties": {
					"id": {"type": "string", "description": ""}
				  }
				}
			  }
			}
		  ],
		  "tool_choice": "any",
		  "parallel_tool_calls": true
		}`
		assert.JSONEq(t, expectedReq, gotReq)
	})

	t.Run("Should retry on 5xx then succeed", func(t *testing.T) {
		// Given
		var attempts int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attempts, 1)
			if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
				http.NotFound(w, r)
				return
			}
			if atomic.LoadInt32(&attempts) <= 2 {
				http.Error(w, `{"error":"temporary"}`, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
              "choices": [
                { "message": { "role": "assistant", "content": "Hello after retries" } }
              ]
            }`))
		}))
		defer srv.Close()

		cfg := &mistral.Config{
			Verbose:           false,
			MistralAPIBaseURL: srv.URL,
			RetryMaxRetries:   3,
			RetryWaitMin:      1 * time.Millisecond,
			RetryWaitMax:      5 * time.Millisecond,
		}
		c := mistral.NewWithConfig("fake-api-key", cfg)
		ctx := context.Background()
		inputMsgs := []mistral.ChatMessage{mistral.NewUserMessageFromString("Hi!")}

		// When
		res, err := c.ChatCompletion(ctx, mistral.NewChatCompletionRequest("mistral-large", inputMsgs))

		// Then
		assert.NoError(t, err)
		assert.Len(t, res.Choices, 1)
		assert.Equal(t, mistral.NewAssistantMessageFromString("Hello after retries"), res.Choices[0].Message)
		assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	})

	t.Run("Should not retry on 400 and fail immediately", func(t *testing.T) {
		// Given
		var attempts int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attempts, 1)
			if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
				http.NotFound(w, r)
				return
			}
			http.Error(w, `{
				"object": "error",
				"message": {
					"detail": [
						{
							"type": "extra_forbidden",
							"loc": [
								"body",
								"parallel_tool_callss"
							],
							"msg": "Extra inputs are not permitted",
							"input": true
						}
					]
				},
				"type": "invalid_request_error",
				"param": null,
				"code": null
			}`, http.StatusBadRequest)
		}))
		defer srv.Close()

		cfg := &mistral.Config{
			Verbose:           false,
			MistralAPIBaseURL: srv.URL,
			RetryMaxRetries:   5,
			RetryWaitMin:      1 * time.Millisecond,
			RetryWaitMax:      2 * time.Millisecond,
		}
		c := mistral.NewWithConfig("fake-api-key", cfg)
		ctx := context.Background()
		inputMsgs := []mistral.ChatMessage{mistral.NewUserMessageFromString("Hi!")}

		// When
		_, err := c.ChatCompletion(ctx, mistral.NewChatCompletionRequest("mistral-large", inputMsgs))

		// Then
		expectedErr := mistral.NewApiError(
			http.StatusBadRequest,
			map[string]any{
				"object": "error",
				"message": map[string]any{
					"detail": []any{
						map[string]any{
							"type":  "extra_forbidden",
							"loc":   []any{"body", "parallel_tool_callss"},
							"msg":   "Extra inputs are not permitted",
							"input": true,
						},
					},
				},
				"type":  "invalid_request_error",
				"param": nil,
				"code":  nil,
			},
		)
		assert.Error(t, err)
		assert.Equal(t, &expectedErr, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
		assert.Equal(t, "[400] invalid_request_error: extra_forbidden: Extra inputs are not permitted (body.parallel_tool_callss)", err.Error())
	})

	t.Run("Should retry on timeout error then succeed", func(t *testing.T) {
		// Given
		successJSON := []byte(`{"choices":[{"message":{"role":"assistant","content":"OK after timeout"}}]}`)
		cfg := &mistral.Config{
			Verbose:           false,
			RetryMaxRetries:   3,
			RetryWaitMin:      1 * time.Millisecond,
			RetryWaitMax:      5 * time.Millisecond,
			MistralAPIBaseURL: "http://invalid.local",
			Transport: &flakyRoundTripper{
				failuresLeft: 1,
				successBody:  successJSON,
			},
			ClientTimeout: 2 * time.Second,
		}
		c := mistral.NewWithConfig("fake-api-key", cfg)
		ctx := context.Background()
		inputMsgs := []mistral.ChatMessage{mistral.NewUserMessageFromString("Hello")}

		// When
		res, err := c.ChatCompletion(ctx, mistral.NewChatCompletionRequest("mistral-large", inputMsgs))

		// Then
		assert.NoError(t, err)
		assert.Len(t, res.Choices, 1)
		assert.Equal(t, mistral.NewAssistantMessageFromString("OK after timeout"), res.Choices[0].Message)
	})

	t.Run("Should fail when max retries reached", func(t *testing.T) {
		// Given
		var attempts int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attempts, 1)
			http.Error(w, `{"error":"unavailable"}`, http.StatusServiceUnavailable)
		}))
		defer srv.Close()

		cfg := &mistral.Config{
			Verbose:           false,
			MistralAPIBaseURL: srv.URL,
			RetryMaxRetries:   2,
			RetryWaitMin:      1 * time.Millisecond,
			RetryWaitMax:      2 * time.Millisecond,
		}
		c := mistral.NewWithConfig("fake-api-key", cfg)
		ctx := context.Background()
		inputMsgs := []mistral.ChatMessage{mistral.NewUserMessageFromString("Hi")}

		// When
		_, err := c.ChatCompletion(ctx, mistral.NewChatCompletionRequest("mistral/mistral-large", inputMsgs))

		// Then
		assert.Error(t, err)
		assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	})
}

func TestChatCompletionResponse(t *testing.T) {
	t.Run("should unmarshall with correct created_at time format", func(t *testing.T) {
		j := `{
			"id": "12345",
			"created": 1764278339,
			"model": "mistral-small-latest",
			"usage": {
				"prompt_tokens": 13,
				"total_tokens": 23,
				"completion_tokens": 10
			},
			"object": "chat.completion",
			"choices": [
				{
					"index": 0,
					"finish_reason": "stop",
					"message": {
						"role": "assistant",
						"tool_calls": null,
						"content": "Hello! How can I assist you today?"
					}
				}
			]
		}`

		var tc mistral.ChatCompletionResponse

		assert.NoError(t, json.Unmarshal([]byte(j), &tc))
		assert.Equal(t, time.Date(2025, time.November, 27, 21, 18, 59, 0, time.UTC), tc.Created)
	})
}
