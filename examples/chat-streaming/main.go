package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/thomas-marquis/mistral-client/mistral"
)

func main() {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		panic("Please set MISTRAL_API_KEY environment variable")
	}

	client := mistral.New(apiKey, mistral.WithClientTimeout(25*time.Second))

	ctx := context.Background()
	req := mistral.NewChatCompletionRequest("mistral-small-latest",
		[]mistral.ChatMessage{
			mistral.NewUserMessageFromString("Tell me a joke."),
		},
		mistral.WithStreaming())
	resChan, err := client.ChatCompletionStream(ctx, req)
	if err != nil {
		panic(err)
	}

	for evt := range resChan {
		if evt.Error != nil {
			panic(evt.Error)
		}
		delta := evt.DeltaMessage()
		fmt.Printf("%s (%fms)\n",
			delta.Content().String(),
			evt.ChunkLatency.Seconds()*1000)
		if evt.IsLastChunk {
			fmt.Printf("Finished (%s)\nIn/Out/Total tokens: %d/%d/%d\nTotal latency: %fms",
				evt.Choices[0].FinishReason,
				evt.Usage.PromptTokens, evt.Usage.CompletionTokens, evt.Usage.TotalTokens,
				evt.TotalLatency.Seconds()*1000)
			break
		}
	}
}
