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

	models, err := client.ListModels(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println("All models:")
	for _, model := range models {
		fmt.Printf("%+v\n", model)
	}

	filtered, err := client.SearchModels(context.Background(), &mistral.ModelCapabilities{
		Moderation: true,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Filtered models:")
	for _, model := range filtered {
		fmt.Printf("%+v\n", model)
	}
}
