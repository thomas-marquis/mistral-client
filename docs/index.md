# mistral-client

HTTP client for Mistral AI written in Go.

<figure markdown="span">
  ![Logo](assets/images/logo-tr.png)
</figure>

## Requirements

- Go 1.25 or higher

## Installation

```bash
go get github.com/thomas-marquis/mistral-client
```

## Features

**Basic features:**

- **Chat Completion**: Synchronous and streaming support.
- **Embeddings**: Generate text embeddings with various encoding formats and dimensions.
- **Tool Calling**: Native support for function calling and tool usage.
- **Multi-modal Input**: Handle images, audio, and documents in your messages.
- **Structured Output**: Support for JSON Mode and JSON Schema.
- **Advanced Client**: Built-in retry logic, rate limiting, and custom HTTP client configuration.
- **Model Management**: List, search, and retrieve details for Mistral models.

**What makes a difference:**

- **Caching**: Cache responses to avoid unnecessary repeated API calls (e.g. for local development runs).

**Coming soon:**

- **MLflow integration**: Store your prompts on MLfow Prompt Registry and get them back for reuse.
- **Fake models**: Use fake models for local development and testing.

## Getting Started

To start using the Mistral client, you first need to create an instance of the client with your API key:

```go
import "github.com/thomas-marquis/mistral-client/mistral"

client := mistral.New(apiKey)
```

You can also customize the client with various options:

```go
client := mistral.New(apiKey,
    mistral.WithClientTimeout(60*time.Second),
    mistral.WithRetry(4, 1*time.Second, 3*time.Second),
)
```

## Basic Usage

### Chat Completion

```go
req := mistral.NewChatCompletionRequest("mistral-small-latest", []mistral.ChatMessage{
    mistral.NewUserMessageFromString("Hello! How are you today?"),
})

res, err := client.ChatCompletion(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Println(res.AssistantMessage().Content().String())
```

### Tool Calling

```go
tools := []mistral.Tool{
    mistral.NewTool("get_weather", "Get the weather for a location", mistral.PropertyDefinition{
        Type: "object",
        Properties: map[string]mistral.PropertyDefinition{
            "location": {Type: "string", Description: "The city and state, e.g. San Francisco, CA"},
        },
    }),
}

req := mistral.NewChatCompletionRequest("mistral-small-latest", []mistral.ChatMessage{
    mistral.NewUserMessageFromString("What's the weather in Paris?"),
}, mistral.WithTools(tools))

res, err := client.ChatCompletion(context.Background(), req)
```

### Embeddings

```go
req := mistral.NewEmbeddingRequest("mistral-embed", []string{"Mistral AI is awesome!"})

res, err := client.Embeddings(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

for _, vector := range res.Embeddings() {
    fmt.Println(vector)
}
```

## Examples

You can find more detailed examples in the `examples` folder:

- [Chat Completion](https://github.com/thomas-marquis/mistral-client/tree/main/examples/chat-completion/main.go): Basic usage of the chat completion API.
- [Chat Completion (Advanced)](https://github.com/thomas-marquis/mistral-client/tree/main/examples/chat-completion-advanced/main.go): Advanced options like retry, rate limiting, and timeout.
- [Chat completion (with structured output)](https://github.com/thomas-marquis/mistral-client/tree/main/examples/chat-completion-constrained/main.go)
- [Chat Audio](https://github.com/thomas-marquis/mistral-client/tree/main/examples/chat-audio/main.go): Transcribe and interact with audio files.
- [Chat Vision](https://github.com/thomas-marquis/mistral-client/tree/main/examples/chat-vision/main.go): Interact with images.
- [Embeddings](https://github.com/thomas-marquis/mistral-client/tree/main/examples/embedding/main.go): Generate text embeddings.
- [Get Model](https://github.com/thomas-marquis/mistral-client/tree/main/examples/get-model/main.go): Retrieve details for a specific model.
- [List Models](https://github.com/thomas-marquis/mistral-client/tree/main/examples/list-models/main.go): List and search available models.
- [Tools / Function Calling](https://github.com/thomas-marquis/mistral-client/tree/main/examples/tools/main.go): Use tools and function calling.

## Useful Links

- [Go package documentation](https://pkg.go.dev/github.com/thomas-marquis/mistral-client)
- [GitHub Repository](https://github.com/thomas-marquis/mistral-client)
- [Mistral AI API Documentation](https://docs.mistral.ai/)