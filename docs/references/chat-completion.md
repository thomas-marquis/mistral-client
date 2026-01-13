# Chat completion

## Methods

### `ChatCompletion`

Calls the `/v1/chat/completions` endpoint.

**Arguments:**

- `ctx`: `context.Context`
- `req`: `*ChatCompletionRequest`

**Returns:** `(*ChatCompletionResponse, error)`

### `ChatCompletionStream`

Calls the `/v1/chat/completions` endpoint with streaming enabled.

**Arguments:**

- `ctx`: `context.Context`
- `req`: `*ChatCompletionRequest`

**Returns:** `(<-chan *CompletionChunk, error)`

## Request

### `ChatCompletionRequest`

The request object for chat completion. It embeds `CompletionConfig`.

**Fields:**

- `Model` (`string`): ID of the model to use.
- `Messages` (`[]ChatMessage`): The list of messages for the conversation.
- `Tools` (`[]Tool`): A list of tools the model may call.
- `MaxTokens` (`int`): Maximum number of tokens to generate. The token count of your prompt plus `max_tokens` cannot exceed the model's context length.
- `Temperature` (`float64`): Sampling temperature. We recommend between 0.0 and 0.7.
- `TopP` (`float64`): Nucleus sampling. Default to 1.0.
- `ResponseFormat` (`*ResponseFormat`): Specifies the format of the output (Text, JSON Object, or JSON Schema).
- `ToolChoice` (`ToolChoiceType`): Controls which (if any) tool is called by the model.
- `ParallelToolCalls` (`bool`): Whether to enable parallel function calling. Default to `true`.
- `FrequencyPenalty` (`float64`): Penalizes repetition based on frequency.
- `PresencePenalty` (`float64`): Penalizes repetition based on presence.
- `N` (`int`): Number of completions to return for each request.
- `PromptMode` (`string`): Toggles between the reasoning mode and no system prompt.
- `RandomSeed` (`int`): Seed for deterministic results.
- `SafePrompt` (`bool`): Whether to inject a safety prompt before all conversations.
- `Stop` (`[]string`): Stop generation if these tokens are detected.
- `Stream` (`bool`): Whether to stream back partial progress.

### `NewChatCompletionRequest`

Creates a new `ChatCompletionRequest`.

**Arguments:**

- `model`: `string`
- `messages`: `[]ChatMessage`
- `...opts`: `ChatCompletionRequestOption`

### `NewChatCompletionStreamRequest`

Creates a new `ChatCompletionRequest` with streaming enabled.

**Arguments:**

- `model`: `string`
- `messages`: `[]ChatMessage`
- `...opts`: `ChatCompletionRequestOption`

## Options

### `WithResponseTextFormat`

Ensures the response is formatted as text. This is the default format.

**Arguments:** none

### `WithResponseJsonSchema`

Ensures the response follows the specified JSON schema.

**Arguments:** `PropertyDefinition`

### `WithResponseJsonObjectFormat`

Ensures the response is a JSON object.

**Arguments:** none

### `WithTools`

Enables the model to call the specified tools. Sets `ToolChoice` to `Auto`.

**Arguments:** `[]Tool`

### `WithToolChoice`

Controls which (if any) tool is called by the model.

**Arguments:** `ToolChoiceType`

### `WithStreaming`

Enables streaming back partial progress.

**Arguments:** none

## Response

### `ChatCompletionResponse`

**Fields:**

- `Choices` (`[]ChatCompletionChoice`): List of completions.
- `Created` (`time.Time`): Creation timestamp.
- `Id` (`string`): Unique ID for the completion.
- `Model` (`string`): Model used for the completion.
- `Object` (`string`): Object type.
- `Usage` (`UsageInfo`): Token usage information.
- `Latency` (`time.Duration`): Request latency.

### `AssistantMessage`

Method on `ChatCompletionResponse` that returns the first assistant message in the choices, or `nil` if there are no assistant messages.

### `ChatCompletionChoice`

**Fields:**

- `FinishReason` (`FinishReason`): The reason why the model stopped generating tokens.
- `Index` (`int`): Index of the choice.
- `Message` (`*AssistantMessage`): The generated assistant message.

### `FinishReason`

Possible values:

- `FinishReasonStop`: Model reached a natural stop point or a provided stop sequence.
- `FinishReasonLength`: Model reached the maximum number of tokens.
- `FinishReasonModelLength`: Model reached its maximum context length.
- `FinishReasonError`: An error occurred during generation.
- `FinishReasonToolCalls`: Model is calling a tool.

## Streaming

### `CompletionChunk`

Represents a chunk of data received during streaming.

**Fields:**

- `Choices` (`[]CompletionResponseStreamChoice`): List of choices in this chunk.
- `Created` (`time.Time`): Creation timestamp.
- `Id` (`string`): Unique ID.
- `Model` (`string`): Model used.
- `Object` (`string`): Object type.
- `Usage` (`UsageInfo`): Token usage (usually only present in the last chunk).
- `IsLastChunk` (`bool`): Indicates if this is the last chunk.
- `ChunkLatency` (`time.Duration`): Latency of this specific chunk.
- `TotalLatency` (`time.Duration`): Total latency of the request.
- `Error` (`error`): Error if any occurred during streaming.

### `DeltaMessage`

Method on `CompletionChunk` that returns the `AssistantMessage` delta for the first choice.
