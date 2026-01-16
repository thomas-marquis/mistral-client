# mistral-client

HTTP client for Mistral AI written in Go. 🚀

[![Go Reference](https://pkg.go.dev/badge/github.com/thomas-marquis/mistral-client.svg)](https://pkg.go.dev/github.com/thomas-marquis/mistral-client)
[![CI](https://github.com/thomas-marquis/mistral-client/actions/workflows/ci.yaml/badge.svg)](https://github.com/thomas-marquis/mistral-client/actions/workflows/ci.yaml)
[![License](https://img.shields.io/github/license/thomas-marquis/mistral-client)](LICENSE)

<p align="center">
  <img src="docs/assets/images/logo-tr.png" width="200" alt="mistral-client logo">
</p>

## ✨ Features

- **Chat Completion**: Synchronous and streaming support.
- **Embeddings**: Generate text embeddings with various encoding formats and dimensions.
- **Tool Calling**: Native support for function calling and tool usage.
- **Multi-modal Input**: Handle images, audio, and documents in your messages.
- **Structured Output**: Support for JSON Mode and JSON Schema.
- **Advanced Client**: Built-in retry logic, rate limiting, and custom HTTP client configuration.
- **Model Management**: List, search, and retrieve details for Mistral models.
- **Caching**: Cache responses to avoid unnecessary repeated API calls (e.g. for local development runs).

## 📦 Installation

### Requirements

- Go 1.25 or higher

### Installation process

```bash
go get github.com/thomas-marquis/mistral-client
```

## 📚 Documentation

- [Project Documentation](https://thomas-marquis.github.io/mistral-client/)
- [Go Package Documentation](https://pkg.go.dev/github.com/thomas-marquis/mistral-client)
- [Mistral AI API Documentation](https://docs.mistral.ai/)

## 💻 Usage

### 🔧 Client Initialization

```go
import "github.com/thomas-marquis/mistral-client/mistral"

client := mistral.New(apiKey)
```

### 💬 Chat Completion

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

### 🛠️ Tool Calling

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

### 🔢 Embeddings

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

## 🤝 Contribute

All contributions are welcome! Feel free to open an issue or submit a PR. ✨

Check out [CONTRIBUTING.md](CONTRIBUTING.md) for more details.
