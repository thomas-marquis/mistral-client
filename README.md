# mistral-client

HTTP client for Mistral AI written in Go

## Usage

Examples are available in the [examples](./examples) directory.

### Client

Create a default client instance:

```go
c := mistral.New(apiKey)
```

Available options:

```go
c := mistral.New(apiKey,
    mistral.WithClientTimeout(60*time.Second)),
	mistral.WithBaseAPIURL(mistral.BaseApiUrl), // default to "https://api.mistral.ai"
    mistral.WithRateLimiter(rate.NewLimiter(rate.Every(1*time.Second), 50)), // uses the package "golang.org/x/time/rate"
	mistral.WithVerbose(true), // Default to false
	mistral.WithRetry(10, 1*time.Second, 5*time.Second),
	mistral.WithRetryStatusCodes(http.StatusTooManyRequests, http.StatusInternalServerError),
)
```

[Example](./examples/chat-completion-advanced/main.go)

### ChatCompletion

From a client instance, build a request and call the ChatCompletion method:

```go
req := mistral.NewChatCompletionRequest(
    "mistral-small-latest", // Specify a model name with chat completion capabilities
    []mistral.ChatMessage{
        mistral.NewSystemMessageFromString(systemPrompt),
        mistral.NewUserMessageFromString(userPrompt),
    })
res, err := client.ChatCompletion(context.Background(), req)
```

[Example](./examples/chat-completion/main.go)

### Embeddings

Similar to ChatCompletion, but with an EmbeddingsRequest instance as input of the Embeddings method:

```go
exts := []string{
    "ipsum eiusmod",
    "dolor sit amet",
}
req := mistral.NewEmbeddingRequest("mistral-embed", texts)
res, err := client.Embeddings(context.Background(), req)
```

[Example](./examples/embedding/main.go)

### ListModels and SearchModels

List all the available models:

```go
models, err := client.ListModels(context.Background())
```

You can also filter models by capabilities:

```go
filtered, err := client.SearchModels(context.Background(), &mistral.ModelCapabilities{
    CompletionChat: true,
    FunctionCalling: true,
})
```

Both methods retuns a list of [`BaseModelCard`](./mistral/models.go) objects.

[Example](./examples/list-models/main.go)

### Get a specific model

Get a model card by its name:

```go
model, err := client.GetModel(context.Background(), "model-not-found")
if err != nil && errors.Is(err, mistral.ErrModelNotFound)  {
    fmt.Println("Model not found")
}
```

[Example](./examples/get-model/main.go)

## Contribute

All contributions are welcome!

- Use golangci-lint before commiting ([install it first](https://golangci-lint.run/docs/welcome/install/local/))
- Open an issue or submit a PR
