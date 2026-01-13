# Embed a text


Embeddings are numerical representations of text that can be used for tasks like semantic search, clustering, or classification.

## Create embeddings

To create embeddings, use the `Embeddings` method from a `Client` instance.
This method expects a context and an `EmbeddingRequest`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/thomas-marquis/mistral-client/mistral"
)

func main() {
	client := mistral.New("API_KEY")

	texts := []string{
		"Hello, how are you?",
		"I'm fine, thank you!",
	}

	req := mistral.NewEmbeddingRequest("mistral-embed", texts)
	res, err := client.Embeddings(context.Background(), req)
	if err != nil {
		panic(err)
	}

	for _, v := range res.Embeddings() { // (1)
		fmt.Printf("Embedding: %v\n", v)
	}
}
```

1. Iterate over the two texts embeddings.

## Available options

You can customize the embedding request using the following options:

- `WithEmbeddingOutputDtype`: Specify the output data type (e.g., `float`, `int8`, `uint8`, `binary`, `ubinary`).
- `WithEmbeddingOutputDimension`: Specify the dimension of the output embeddings (if supported by the model).
- `WithEmbeddingEncodingFormat`: Specify the encoding format (e.g., `float`, `base64`).

```go
req := mistral.NewEmbeddingRequest("mistral-embed", texts,
    mistral.WithEmbeddingOutputDtype(mistral.EmbeddingOutputDtypeFloat),
    mistral.WithEmbeddingOutputDimension(512), // (1)
)
```

1. Keep in mind that not all models support custom dimensions.

## Links

- [Mistral's API documentation](https://docs.mistral.ai/api/endpoint/embeddings#operation-embeddings_v1_embeddings_post)