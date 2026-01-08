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

Then you can use it to call the Mistral API. For example, to create a chat completion:

```go
req := mistral.NewChatCompletionRequest(
    "mistral-small-latest",
    []mistral.ChatMessage{
        mistral.NewUserMessageFromString("Hello, how are you?"),
    })
res, err := client.ChatCompletion(context.Background(), req)
if err != nil {
    // handle error
}

fmt.Println(res.AssistantMessage().MessageContent)
```

## Examples

You can find more detailed examples in the `examples` folder:

- [Chat Completion](https://github.com/thomas-marquis/mistral-client/tree/main/examples/chat-completion): Basic usage of the chat completion API.
- [Chat Completion (Advanced)](https://github.com/thomas-marquis/mistral-client/tree/main/examples/chat-completion-advanced): Advanced options like retry, rate limiting, and timeout.
- [Chat Audio](https://github.com/thomas-marquis/mistral-client/tree/main/examples/chat-audio): Transcribe and interact with audio files.
- [Chat Vision](https://github.com/thomas-marquis/mistral-client/tree/main/examples/chat-vision): Interact with images.
- [Embeddings](https://github.com/thomas-marquis/mistral-client/tree/main/examples/embedding): Generate text embeddings.
- [Get Model](https://github.com/thomas-marquis/mistral-client/tree/main/examples/get-model): Retrieve details for a specific model.
- [List Models](https://github.com/thomas-marquis/mistral-client/tree/main/examples/list-models): List and search available models.
- [Tools / Function Calling](https://github.com/thomas-marquis/mistral-client/tree/main/examples/tools): Use tools and function calling.

## Useful Links

- [GitHub Repository](https://github.com/thomas-marquis/mistral-client)
- [Mistral AI API Documentation](https://docs.mistral.ai/)