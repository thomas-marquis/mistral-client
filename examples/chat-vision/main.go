package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/thomas-marquis/mistral-client/mistral"
)

func main() {
	const imageUrl = "https://images.radio-canada.ca/v1/ici-info/16x9/lni-ligue-nationale-improvisation-3.jpg"

	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		panic("Please set MISTRAL_API_KEY environment variable")
	}
	client := mistral.New(apiKey,
		mistral.WithClientTimeout(60*time.Second))

	userPrompt := "How many people are in the image?"

	req := mistral.NewChatCompletionRequest(
		"mistral-small-latest",
		[]mistral.ChatMessage{
			mistral.NewUserMessage(mistral.ContentChunks{
				mistral.NewImageUrlContent(imageUrl),
				mistral.NewTextContent(userPrompt),
			}),
		},
		mistral.WithResponseJsonSchema(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"count": map[string]any{
					"type": "integer",
				},
			},
		}))
	res, err := client.ChatCompletion(context.Background(), req)
	if err != nil {
		panic(err)
	}

	msg := res.AssistantMessage()
	if msg != nil {
		fmt.Println(msg.MessageContent)
	} else {
		panic("No assistant message found")
	}
}
