package main

import (
	"context"
	"fmt"
	"os"

	"github.com/thomas-marquis/mistral-client/mistral"
)

func main() {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		panic("Please set MISTRAL_API_KEY environment variable")
	}
	client := mistral.New(apiKey)

	// with response json object format
	req := mistral.NewChatCompletionRequest(
		"mistral-small-latest",
		[]mistral.ChatMessage{
			mistral.NewUserMessageFromString("Tell me a joke about dogs. The output must contains a 'joke' filed."),
		},
		mistral.WithResponseJsonObjectFormat())
	res, err := client.ChatCompletion(context.Background(), req)
	if err != nil {
		panic(err)
	}

	msg := res.AssistantMessage()

	var joke map[string]any
	if err := msg.Output(&joke); err != nil {
		panic(err)
	}

	fmt.Println(joke)

	// with response json schema
	req = mistral.NewChatCompletionRequest(
		"mistral-small-latest",
		[]mistral.ChatMessage{
			mistral.NewUserMessageFromString("Tell me a joke about cats."),
		},
		mistral.WithResponseJsonSchema(mistral.NewObjectPropertyDefinition(
			map[string]mistral.PropertyDefinition{
				"joke": {Type: "string", Description: "A joke."},
				"laugh_level": {
					Type:        "integer",
					Description: "How funny is the joke? In percentage.",
					Default:     50},
			},
		)))
	res, err = client.ChatCompletion(context.Background(), req)
	if err != nil {
		panic(err)
	}

	msg = res.AssistantMessage()
	if err := msg.Output(&joke); err != nil {
		panic(err)
	}

	fmt.Println(joke)
}
