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

	// Initialize the client with local cache enabled.
	// By default, it uses "./mistral/cache" directory.
	// You can also use WithCacheDir("your/custom/path") to specify a different directory.
	client := mistral.New(apiKey,
		mistral.WithLocalCache(),
		mistral.WithClientTimeout(60*time.Second))

	ctx := context.Background()

	// 1. Chat Completion Example
	fmt.Println("--- Chat Completion ---")
	chatReq := mistral.NewChatCompletionRequest(
		"mistral-small-latest",
		[]mistral.ChatMessage{
			mistral.NewUserMessageFromString("Explain what is a cache in one sentence."),
		})

	start := time.Now()
	res, err := client.ChatCompletion(ctx, chatReq)
	if err != nil {
		panic(err)
	}
	fmt.Printf("First call latency: %v\n", time.Since(start))
	fmt.Printf("Response: %s\n\n", res.AssistantMessage().MessageContent)

	// Second call with same request should be almost instantaneous due to cache
	start = time.Now()
	res, err = client.ChatCompletion(ctx, chatReq)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Second call (cached) latency: %v\n", time.Since(start))
	fmt.Printf("Response: %s\n\n", res.AssistantMessage().MessageContent)

	// 2. Embedding Example
	fmt.Println("--- Embedding ---")
	embedReq := mistral.NewEmbeddingRequest("mistral-embed", []string{"Mistral AI is awesome"})

	start = time.Now()
	embedRes, err := client.Embeddings(ctx, embedReq)
	if err != nil {
		panic(err)
	}
	fmt.Printf("First call latency: %v\n", time.Since(start))
	fmt.Printf("Embeddings length: %d\n\n", len(embedRes.Data[0].Embedding))

	// Second call with same request should be almost instantaneous due to cache
	start = time.Now()
	embedRes, err = client.Embeddings(ctx, embedReq)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Second call (cached) latency: %v\n", time.Since(start))
	fmt.Printf("Embeddings length: %d\n", len(embedRes.Data[0].Embedding))
}
