# Perform a basic chat completion

Chat completion consists of exchanging messages between a user and an assistant. 
Each message is composed of a role (system, user, assistant or response) and a content.

We send a list of messages to the API and receive the next message from the assistant.

**Important**: The messages order matters. The API expects that:
- the first message is a user or a system message
- the last message is a user or a tool message

## Sending a request

The first step to perform a chat completion is to create a request:

```go
req := mistral.NewChatCompletionRequest("mistral-small-latest", // (1)
   []mistral.ChatMessage{
   mistral.NewUserMessageFromString("Tell me a joke about cats"),
}) // (2)
```

1. The model to use. You need a model with the `CompletionChat` capability.
2. A list of messages to send. You have to respect the order of the messages (\[system/\]user/assistant/...).


Then, call the `ChatCompletion` from your client's instance method with the request:

```go
client := mistral.New("<your_api_key>")

res, err := client.ChatCompletion(ctx, req)
```

Check there is no error and get the assistant message:

```go
if err != nil {
	// handle error
}
am := res.AssistantMessage() // (1)
```

1. To get the assistant message, you can either use the `AssistantMessage` method or the `Chices[0].Message` attribute.
   It's a pointer.


Complete example:

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
	req := mistral.NewChatCompletionRequest("mistral-small-latest",
		[]mistral.ChatMessage{
			mistral.NewUserMessageFromString("Tell me a joke about cats"),
		})
	res, err := client.ChatCompletion(ctx, req)
	if err != nil {
		panic(err)
	}

	am := res.AssistantMessage()
	fmt.Println(am.Content().String())
}
```

## Customize the request

As shown above, the request can be created with the `NewChatCompletionRequest` function.
It returns a `mistral.ChatCompletionRequest` object. 

This function takes at least two arguments:

-  The model name. The choosen model **must** have at least the `CompletionChat` capability.
   See [here](models.md){ data-preview } for more information about capabilities.
-  The list of messages to send. This list **must** respect the order of the messages (\[system/\]user/assistant/...). System message is optional.
   See below for more information about messages.

You can customize the request either by using options (`ChatCompletionRequestOption`) or setting the request's exported fields directly.

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

req.MaxTokens = 100
```

See available options [here](../references/chat-completion.md#options)
