# Init client

The very first step to use this library is to create a `Client` instance.

To do so, you need an API key (you can create one [here](https://docs.mistral.ai/getting-started/quickstart))

```go
apiKey := os.Getenv("MISTRAL_API_KEY")

client := mistral.New(apiKey)
```

You can tune your client with some options. For example:

```go
apiKey := os.Getenv("MISTRAL_API_KEY")

client := mistral.New(apiKey)
```

The complete list of available is available [here](../references/client.md#available-options).