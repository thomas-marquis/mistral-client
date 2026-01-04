package main

import (
	"context"
	"errors"
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

	model, err := client.GetModel(context.Background(), "mistral-medium-latest")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s (%s): %+v\n", model.Id, model.Description, model.Capabilities)

	_, err = client.GetModel(context.Background(), "model-not-found")
	if err != nil {
		if errors.Is(err, mistral.ErrModelNotFound) {
			fmt.Println("Model not found")
			return
		}
		panic(err)
	}
}
