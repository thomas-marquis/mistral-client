package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/thomas-marquis/mistral-client/mistral"
)

func main() {
	const filePath = "examples/chat-audio/sample.wav"

	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	base64Str := base64.StdEncoding.EncodeToString(data)

	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		panic("Please set MISTRAL_API_KEY environment variable")
	}
	client := mistral.New(apiKey,
		mistral.WithClientTimeout(60*time.Second))

	userPrompt := "Transcribe the provided audio file."

	req := mistral.NewChatCompletionRequest(
		"voxtral-small-latest",
		[]mistral.ChatMessage{
			mistral.NewUserMessage(mistral.ContentChunks{
				mistral.NewAudioContent(base64Str),
				mistral.NewTextContent(userPrompt),
			}),
		})
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
