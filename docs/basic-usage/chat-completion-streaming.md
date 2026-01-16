# Streaming chat completion

Streaming chat completion allows you to receive the assistant's response in real-time, as it is being generated. This is particularly useful for building interactive applications like chatbots where you want to show the text to the user as soon as possible.

## Sending a streaming request

To perform a streaming chat completion, you use the `ChatCompletionStream` method. The request is created similarly to a standard chat completion, but usually with the `NewChatCompletionStreamRequest` helper which sets the `Stream` parameter to `true`.

```go
req := mistral.NewChatCompletionStreamRequest("mistral-small-latest",
    []mistral.ChatMessage{
        mistral.NewUserMessageFromString("Tell me a long story about a space hamster"),
    })

resChan, err := client.ChatCompletionStream(ctx, req)
if err != nil {
    // handle error
}
```

## Handling the stream

The `ChatCompletionStream` method returns a channel of `*mistral.CompletionChunk`. You should iterate over this channel to receive the partial responses (chunks).

Each chunk contains a `Delta` message, which represents the new text generated since the last chunk.

```go
for evt := range resChan {
    if evt.Error != nil {
        // handle error within the stream
        break
    }

    // Get the partial message from the chunk
    delta := evt.DeltaMessage()
    if delta.Content() != nil {
        fmt.Print(delta.Content().String())
    }

    // Check if it's the last chunk
    if evt.IsLastChunk {
        fmt.Printf("\n\nFinished with reason: %s\n", evt.Choices[0].FinishReason)
        fmt.Printf("Tokens: %d (prompt) + %d (completion) = %d (total)\n",
            evt.Usage.PromptTokens, evt.Usage.CompletionTokens, evt.Usage.TotalTokens)
    }
}
```

### Key properties of `CompletionChunk`

- `DeltaMessage()`: Returns the `ChatMessage` containing the new content in this chunk.
- `IsLastChunk`: A boolean indicating if this is the final chunk of the stream.
- `Usage`: On the last chunk, this contains the token usage information.
- `Error`: If an error occurs during streaming, it will be populated in this field.
- `ChunkLatency`: The time it took to receive this specific chunk.
- `TotalLatency`: On the last chunk, the total time elapsed since the request started.

## Complete example

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/thomas-marquis/mistral-client/mistral"
)

func main() {
	client := mistral.New(os.Getenv("MISTRAL_API_KEY"))

	ctx := context.Background()
	req := mistral.NewChatCompletionStreamRequest("mistral-small-latest",
		[]mistral.ChatMessage{
			mistral.NewUserMessageFromString("Tell me a joke."),
		})

	resChan, err := client.ChatCompletionStream(ctx, req)
	if err != nil {
		panic(err)
	}

	for evt := range resChan {
		if evt.Error != nil {
			panic(evt.Error)
		}

		delta := evt.DeltaMessage()
		if delta.Content() != nil {
			fmt.Print(delta.Content().String())
		}

		if evt.IsLastChunk {
			fmt.Printf("\n\n(Finished with reason: %s)\n", evt.Choices[0].FinishReason)
			break
		}
	}
}
```
