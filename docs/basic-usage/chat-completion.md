# Chat completion

Chat completion is the most basic use case for an AI client.
Each call needs (at least) a message list as input and returns the assistant message in the response.

Call the `ChatCompltion` method from a `Client` instance.

This method expects two arguments: a context and a request (`ChatCompletionRequest`).
You can build the request with the function `NewChatCompletionRequest`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/thomas-marquis/mistral-client/mistral"
)

func main() {
	client := mistral.New("API_KEY")

	ctx := context.Background()
	req := mistral.NewChatCompletionRequest("mistral-small-latest", // (1)
		[]mistral.ChatMessage{
			mistral.NewUserMessageFromString("Tell me a joke about cats"),
		}) // (2)
	res, err := client.ChatCompletion(ctx, req)
	if err != nil {
		panic(err)
	}

	fmt.Println(res.AssistantMessage().Content().String()) // (3)
}
```

1. The model to use. You need a model with the `CompletionChat` capability.
2. A list of messages to send. You have to respect the order of the messages (\[system/\]user/assistant/...).
3. To get the assistant message, you can either use the `AssistantMessage` method or the `Chices[0].Message` attribute.
   It's a pointer.

## Creating the request

As shown above, create the request with the `NewChatCompletionRequest` function.

This function takes at least two arguments:

-  The model name. The choosen model **must** have at least the `CompletionChat` capability.
   See [here](models.md){ data-preview } for more information about capabilities.
-  The list of messages to send. This list **must** respect the order of the messages (\[system/\]user/assistant/...). System message is optional.
   See below for more information about messages.

You can then specify a list of options (`ChatCompletionRequestOption`) to customize the request.

```go
req := mistral.NewChatCompletionRequest("mistral-small-latest",
    []mistral.ChatMessage{
        mistral.NewSystemMessageFromString("You're a useful assistant."),
        mistral.NewUserMessageFromString("Tell me a joke"),
        mistral.NewAssistantMessageFromString("What do you call a fake noodle? An impasta!"),
		mistral.NewUserMessage(
			mistral.NewTextChunk("please describe this picture")),
			mistral.NewImageChunk("https://example.com/image.jpg"),
	})
}
```

See available options [here](../references/chat-completion.md#options)

### Specify a response format

- **`WithResponseTextFormat`**

It is the default option (you unlikely need to specify it). Simply instruct the model that the expected response format is a text (without any specific structure, unless if it is specified in the prompt).

- **`WithResponseJsonObjectFormat`**

This option instructs the model that the expected response format is a JSON object but without any specific structure.
You can specify the expected structure in your prompt.

- **`WithResponseJsonSchema`**

This option instructs the model that the expected response format is a JSON object with a specific structure.

Example:

```go
req := mistral.NewChatCompletionRequest("mistral-small-latest",
	messages,
	mistral.WithResponseJsonSchema(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"answer": map[string]any{
				"type": "string",
			},
		},
    }),
)
```