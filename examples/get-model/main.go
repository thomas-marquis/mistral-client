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

	model, err := client.GetModel(context.Background(), "mistral-small-latest")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", model)

	_, err = client.GetModel(context.Background(), "model-not-found")
	if err != nil {
		if errors.Is(err, mistral.ModelNotFoundErr) {
			fmt.Println("Model not found")
			return
		}
		panic(err)
	}
}
