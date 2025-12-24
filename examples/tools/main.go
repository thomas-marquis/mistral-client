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

	completionModel := "mistral-small-latest"

	systemPrompt := `You are a senior python developer.
You're experienced into developing complex and modular application following the clean architecture principles.
Call the tool write_file to write the code the user ask your to write.`

	userPrompt := `Write a simple and nice todo list tkinter application in a single file named main.py.`

	req := mistral.NewChatCompletionRequest(
		completionModel,
		[]mistral.ChatMessage{
			mistral.NewSystemMessageFromString(systemPrompt),
			mistral.NewUserMessageFromString(userPrompt),
		}, mistral.WithTools([]mistral.Tool{
			mistral.NewTool(
				"write_file",
				"Write a file with the given content",
				map[string]any{
					"type": "object",
					"properties": map[string]any{
						"content": map[string]any{
							"type":        "string",
							"description": "The content to write in the file",
						},
						"filename": map[string]any{
							"type":        "string",
							"description": "The filename with extension to write in the current working directory",
						},
					},
				},
			),
		}))
	req.MaxTokens = 128_000
	res, err := client.ChatCompletion(context.Background(), req)
	if err != nil {
		panic(err)
	}

	msg := res.AssistantMessage()
	if msg != nil {
		fmt.Printf("Content:\n%s\n", msg.Content)
		if len(msg.ToolCalls) > 0 {
			for _, call := range msg.ToolCalls {
				fmt.Printf("Function %s called with arguments:\n%+v\n", call.Function.Name, call.Function.Arguments)
			}
		}
	} else {
		panic("No assistant message found")
	}

	fmt.Printf("Latency: %fs\n", res.Latency.Seconds())
	fmt.Printf("Input tokens: %d; Completion tokens: %d; Total tokens: %d\n",
		res.Usage.PromptTokens, res.Usage.CompletionTokens, res.Usage.TotalTokens)
}
