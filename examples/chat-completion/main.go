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
Take a deep breath and proceed step by step:
- first think about the folder structure
- then implement each file`

	userPrompt := `Write a simple and nice todo list tkinter application in a single file named main.py.
Specifications:
- The user may be able to create a task, mark it as done, rename it or delete it.
- The look and feel of the application must follows the Material design guidelines.
- The full application must also be usable with teh keyboard only
- Implement a simple task list system with a default task list provided
- All the content (tasks and list) must be persisted locally in a json file
`

	req := mistral.NewChatCompletionRequest(
		completionModel,
		[]mistral.ChatMessage{
			mistral.NewSystemMessageFromString(systemPrompt),
			mistral.NewUserMessageFromString(userPrompt),
		})
	req.MaxTokens = 128_000
	res, err := client.ChatCompletion(context.Background(), req)
	if err != nil {
		panic(err)
	}

	msg := res.AssistantMessage()
	if msg != nil {
		fmt.Println(msg.Content)
	} else {
		panic("No assistant message found")
	}

	fmt.Printf("Latency: %fs\n", res.Latency.Seconds())
	fmt.Printf("Input tokens: %d; Completion tokens: %d; Total tokens: %d\n",
		res.Usage.PromptTokens, res.Usage.CompletionTokens, res.Usage.TotalTokens)
}
