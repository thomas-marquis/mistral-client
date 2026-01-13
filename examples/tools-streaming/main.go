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

	completionModel := "mistral-large-latest"

	systemPrompt := `You are a senior python developer.
You're experienced into developing complex and modular application following the clean architecture principles.
Call the tool write_file to write the code the user ask your to write.`

	userPrompt := `Write a simple and nice todo list tkinter application.
Split the application code in multiple files and write them.`

	req := mistral.NewChatCompletionStreamRequest(
		completionModel,
		[]mistral.ChatMessage{
			mistral.NewSystemMessageFromString(systemPrompt),
			mistral.NewUserMessageFromString(userPrompt),
		}, mistral.WithTools([]mistral.Tool{
			mistral.NewTool(
				"write_file",
				"Write a file with the given content",
				mistral.NewObjectPropertyDefinition(map[string]mistral.PropertyDefinition{
					"content":  {Type: "string", Description: "The content to write in the file"},
					"filename": {Type: "string", Description: "The filename with extension to write in the current working directory"},
				}),
			),
		}))
	req.MaxTokens = 128_000
	res, err := client.ChatCompletionStream(context.Background(), req)
	if err != nil {
		panic(err)
	}

	for evt := range res {
		fmt.Println(">>>")
		if evt.Error != nil {
			panic(evt.Error)
		}
		delta := evt.DeltaMessage()

		if len(delta.ToolCalls) > 0 {
			for _, call := range delta.ToolCalls {
				fmt.Printf("--\nFunction %s called with arguments:\n%+v\n", call.Function.Name, call.Function.Arguments)
			}
		}

		if evt.IsLastChunk {
			fmt.Printf("Finished (%s)\nIn/Out/Total tokens: %d/%d/%d\nTotal latency: %fms",
				evt.Choices[0].FinishReason,
				evt.Usage.PromptTokens, evt.Usage.CompletionTokens, evt.Usage.TotalTokens,
				evt.TotalLatency.Seconds()*1000)
			break
		}
	}
}
