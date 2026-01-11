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

func TestClient_ChatCompletion(t *testing.T) {
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
		assert.Equal(t, mistral.RoleAssistant, res.Choices[0].Message.Role())
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
				  "properties": {
					"a": {"type": "number"},
					"b": {"type": "number"}
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
				  "properties": {
					"id": {"type": "string"}
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

		c := mistral.New("fake-api-key",
			mistral.WithBaseApiUrl(srv.URL),
			mistral.WithRetry(3, 1*time.Millisecond, 5*time.Millisecond),
		)
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

		c := mistral.New("fake-api-key",
			mistral.WithBaseApiUrl(srv.URL),
			mistral.WithRetry(5, 1*time.Millisecond, 2*time.Millisecond),
		)
		ctx := context.Background()
		inputMsgs := []mistral.ChatMessage{mistral.NewUserMessageFromString("Hi!")}

		// When
		_, err := c.ChatCompletion(ctx, mistral.NewChatCompletionRequest("mistral-large", inputMsgs))

		// Then
		assert.Error(t, err)
		assert.ErrorAs(t, err, new(mistral.ApiError))
		assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
		assert.Equal(t, "[400] invalid_request_error: extra_forbidden: Extra inputs are not permitted (body.parallel_tool_callss)", err.Error())
		assert.Equal(t, map[string]any{
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
		}, err.(mistral.ApiError).Content())
		assert.Equal(t, http.StatusBadRequest, err.(mistral.ApiError).Code())
	})

	t.Run("Should retry on timeout error then succeed", func(t *testing.T) {
		// Given
		successJSON := []byte(`{"choices":[{"message":{"role":"assistant","content":"OK after timeout"}}]}`)
		c := mistral.New("fake-api-key",
			mistral.WithRetry(3, 1*time.Millisecond, 5*time.Millisecond),
			mistral.WithClientTransport(&flakyRoundTripper{
				failuresLeft: 1,
				successBody:  successJSON,
			}),
			mistral.WithBaseApiUrl("ttp://invalid.local"),
			mistral.WithClientTimeout(2*time.Second),
		)
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

		c := mistral.New("fake-api-key",
			mistral.WithRetry(2, 1*time.Millisecond, 2*time.Millisecond),
			mistral.WithBaseApiUrl(srv.URL),
		)
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

func TestClientImpl_ChatCompletionStream(t *testing.T) {
	t.Run("Should call Mistral /chat/completion endpoint", func(t *testing.T) {
		// Given
		var gotReq string
		mockServer := makeMockSseServerWithCapture(t, "POST", "/v1/chat/completions",
			[]string{
				`data: {"id":"aa","object":"chat.completion.chunk","created":1768084548,"model":"mistral-small-latest","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}`,
				`data: {"id":"aa","object":"chat.completion.chunk","created":1768084548,"model":"mistral-small-latest","choices":[{"index":0,"delta":{"content":"Hello "},"finish_reason":null}]}`,
				`data: {"id":"aa","object":"chat.completion.chunk","created":1768084548,"model":"mistral-small-latest","choices":[{"index":0,"delta":{"content":"world! "},"finish_reason":null}]}`,
				`data: {"id":"aa","object":"chat.completion.chunk","created":1768084548,"model":"mistral-small-latest","choices":[{"index":0,"delta":{"content":""},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"total_tokens":5,"completion_tokens":2}}`,
				`data: [DONE]`,
			},
			http.StatusOK, &gotReq)
		defer mockServer.Close()

		ctx := context.TODO()
		c := mistral.New("fakeApiKey", mistral.WithBaseApiUrl(mockServer.URL))
		inputMsgs := []mistral.ChatMessage{
			mistral.NewSystemMessageFromString("You are a helpful assistant."),
			mistral.NewUserMessageFromString("Hello!"),
		}

		// When
		res, err := c.ChatCompletionStream(ctx, mistral.NewChatCompletionStreamRequest("mistral-small-latest", inputMsgs))

		// Then
		assert.NoError(t, err)

		var chunks []*mistral.CompletionChunk
		for evt := range res {
			chunks = append(chunks, evt)
		}
		assert.Equal(t, 4, len(chunks))

		assert.Equal(t, "", chunks[0].Choices[0].Delta.Content().String())
		assert.Equal(t, "Hello ", chunks[1].Choices[0].Delta.Content().String())
		assert.Equal(t, "world! ", chunks[2].Choices[0].Delta.Content().String())

		lastChunk := chunks[3]
		lastMessage := lastChunk.DeltaMessage()
		assert.Equal(t, "", lastMessage.Content().String())

		assert.Equal(t, 3, lastChunk.Usage.PromptTokens)
		assert.Equal(t, 2, lastChunk.Usage.CompletionTokens)
		assert.Equal(t, 5, lastChunk.Usage.TotalTokens)
		assert.Equal(t, mistral.FinishReasonStop, lastChunk.Choices[0].FinishReason)

		for _, c := range chunks {
			assert.Equal(t, "mistral-small-latest", c.Model)
			assert.Equal(t, "aa", c.Id)
			assert.Equal(t, time.Date(2026, time.January, 10, 22, 35, 48, 0, time.UTC), c.Created)
			assert.Equal(t, "chat.completion.chunk", c.Object)
		}

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
		  "parallel_tool_calls": true,
          "stream": true
		}`
		assert.JSONEq(t, expectedReq, gotReq)
	})
}
