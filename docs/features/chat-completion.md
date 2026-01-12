# Chat completion










## Messages

The interface `ChatMessage` represent a message sent or received by the model.
This interface has four implementations:
- [`SystemMessage`](#system-message)
- [`UserMessage`](#user-message)
- [`AssistantMessage`](#assistant-message)
- [`ToolMessage`](#tool-message)

### System message

With a simple string:
```go
mistral.NewSystemMessageFromString("You're a useful assistant.")
```

### User message


With a simple string:
```go
mistral.NewUserMessageFromString("Tell me a joke")
```

### Assistant message

Assistant messages are returned by the model. You unlikely need to create them manually.
But just in case, here is how to do it:

```go
mistral.NewAssistantMessageFromString("What do you call a fake noodle? An impasta!")
```

### Tool message

Tool messages are emitted after a tool was actually called. 
This type of message contains the tool's response.

Example with a simple string:
```go
mistral.NewToolMessage(
    "toll-name",
	"tool-call-id",
	mistral.ContentString("a string returned by the tool") // (1)
)
```

1. Could be either a content string or a list of chunks.
   See [message content](#message-content) below for more information.

## Message content

There are two types of content for a message:
- a simple string
- chunks

### Simple string

The easiest way to create a simple string message is to use the functions:
- `NewUserMessageFromString`
- `NewSystemMessageFromString`
- `NewAssistantMessageFromString`

### Chunks

If you want to use multimodal features (like images or audio), or if you want to provide more complex content, you can use chunks.

```go
mistral.NewUserMessage(mistral.ContentChunks{
    mistral.NewTextChunk("Describe this image:"),
    mistral.NewImageUrlChunk("https://example.com/image.jpg"),
})
```

Supported chunks:
- `TextChunk`: a simple text block.
- `ImageUrlChunk`: a link to an image.
- `AudioChunk`: a base64 encoded audio string.
- `DocumentUrlChunk`: a link to a document.
- `FileChunk`: a reference to a file uploaded to Mistral.

## Handling the response

The `ChatCompletion` method returns a `ChatCompletionResponse`.
You can use the `AssistantMessage` method to easily get the assistant's response.

```go
res, err := client.ChatCompletion(ctx, req)
if err != nil {
    panic(err)
}

msg := res.AssistantMessage()
fmt.Println(msg.Content().String())
```

If the model used tools, you can check the `ToolCalls` attribute of the assistant message.

```go
if len(msg.ToolCalls) > 0 {
    for _, call := range msg.ToolCalls {
        fmt.Printf("Function %s called with arguments: %v\n", 
            call.Function.Name, call.Function.Arguments)
    }
}
```

## Tools / Function calling

To use tools, you first need to define them and then pass them to the request using the `WithTools` option.

```go
tool := mistral.NewTool("get_weather", "Get the weather in a location", map[string]any{
    "type": "object",
    "properties": map[string]any{
        "location": map[string]any{
            "type": "string",
            "description": "The city and state, e.g. San Francisco, CA",
        },
    },
    "required": []string{"location"},
})

req := mistral.NewChatCompletionRequest("mistral-small-latest",
    messages,
    mistral.WithTools([]mistral.Tool{tool}),
)
```

You can also control how the model uses tools with `WithToolChoice`:

```go
mistral.WithToolChoice(mistral.ToolChoiceAny) // Force use of any tool
mistral.WithToolChoice(mistral.ToolChoiceNone) // Disable tools
mistral.WithToolChoice(mistral.ToolChoiceAuto) // Let the model decide (default)
```

## Links

- [Mistral's API documentation](https://docs.mistral.ai/api/endpoint/chat#operation-chat_completion_v1_chat_completions_post)