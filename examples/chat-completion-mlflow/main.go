package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/thomas-marquis/mistral-client/mistral"
	"github.com/thomas-marquis/mistral-client/mlflow"
)

// This example needs a running-local MLflow server with a chat prompt named "python_dev_chat"

// System prompt:
// You are a senior python developer. You're experienced into developing complex and modular application following the clean architecture principles. Take a deep breath and proceed step by step:
// - first think about the folder structure
// - then implement each file

// User prompt template:
// Write a simple and nice tkinter application in a single file named main.py. Specifications: {{specifications}}

// Check the CONTRIBUTION.md file for instructions on how to start a local MLflow server

func main() {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		panic("Please set MISTRAL_API_KEY environment variable")
	}
	client := mistral.New(apiKey,
		mistral.WithClientTimeout(60*time.Second))

	promptRegistry, err := mlflow.NewPromptRegistry("http://localhost:5000")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	messages, err := mistral.MessagesFromRegisteredPrompt(ctx, promptRegistry,
		"python_dev_chat", mlflow.VersionLatest,
		map[string]any{"specifications": "an image viewer application"})
	if err != nil {
		panic(err)
	}

	req := mistral.NewChatCompletionRequest("mistral-small-latest", messages)
	req.MaxTokens = 128_000

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

	fmt.Printf("Latency: %fs\n", res.Latency.Seconds())
	fmt.Printf("Input tokens: %d; Completion tokens: %d; Total tokens: %d\n",
		res.Usage.PromptTokens, res.Usage.CompletionTokens, res.Usage.TotalTokens)
}
