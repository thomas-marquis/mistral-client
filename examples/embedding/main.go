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
	client := mistral.New(apiKey,
		mistral.WithClientTimeout(60*time.Second))

	texts := []string{
		"ipsum eiusmod",
		"dolor sit amet",
	}

	res, err := client.Embeddings(context.Background(),
		mistral.NewEmbeddingRequest("mistral-embed", texts))
	if err != nil {
		panic(err)
	}
	for _, v := range res.Embeddings() {
		fmt.Printf("Embedding size = %d\n", len(v))
	}

	fmt.Printf("Latency: %fs\n", res.Latency.Seconds())
	fmt.Printf("Input tokens: %d; Completion tokens: %d; Total tokens: %d\n",
		res.Usage.PromptTokens, res.Usage.CompletionTokens, res.Usage.TotalTokens)
}
