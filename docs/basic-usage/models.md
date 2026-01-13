# List and Search models

Explore the Mistral's model catalogue and find information about the models.

## List models

`mistral-client` provides features to list all the models available on the Mistral platform.

```go
package main

import (
    "context"
    "errors"
    "fmt"

    "github.com/thomas-marquis/mistral-client/mistral"
)

func main() {
    client := mistral.New("API_KEY")

    models, err := client.ListModels(context.Background()) // (1)
    if err != nil {
        panic(err)
    }

    for _, model := range models {
        fmt.Printf("Model ID: %s\n", model.Id)
    }
}
```

1. This method lists all the models available on the Mistral platform.
   Each model is represented by the `mistral.BaseModelCard` struct.

## Search models

You can also filter models by their capabilities.

```go
filtered, err := client.SearchModels(context.Background(), &mistral.ModelCapabilities{
    Audio: true, // (1)
})
```

1. This filter will return only models that support audio processing.
   Check the `mistral.ModelCapabilities` struct for all the available options.

## Get model details

To get details about a specific model, use the `GetModel` method.

```go
model, err := client.GetModel(context.Background(), "mistral-medium-latest")
if err != nil {
    if errors.Is(err, mistral.ErrModelNotFound) {
        fmt.Println("Model not found")
        return
    }
    panic(err)
}

fmt.Printf("Model ID: %s, Description: %s\n", model.Id, model.Description)
```

## Links

- [Mistral's API documentation](https://docs.mistral.ai/api/endpoint/models#operation-list_models_v1_models_get)

