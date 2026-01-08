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

1. The model to use.
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

## Available options

- `WithResponseTextFormat`
- `WithResponseJsonSchema`
- `WithResponseJsonObjectFormat`
- `WithTools`
- `WithToolChoice`

### Specify a response format

- **`WithResponseTextFormat`**

It is the default option (you unlikely need to specify it). Simply instruct the model that the expected response format is a text (without any specific structure, unless if it is specified in the prompt).

- **`WithResponseJsonObjectFormat`**

This option instruct the model that the expected response format is a JSON object but without any specific structure.
You can specify the expected struture in your prompt.

- **`WithResponseJsonSchema`**

This option instruct the model that the expected response format is a JSON object with a specific structure.

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

### Specify tools

- **`WithTools`**

Specify a list of tools to use.

A tool is a function that the model may decide to call (or not). 
You can specify multiple tools, and the model can decide to call zero or more of them (even calling a single tool multiple times).

```go
req := mistral.NewChatCompletionRequest("mistral-small-latest",
	messages,
	mistral.WithTools([]mistral.Tool{
	    mistral.NewTool("add", "add two numbers", map[string]any{
		    "type": "object",
			"properties": map[string]any{
				"a": map[string]any{
					"type": "number",
				},
				"b": map[string]any{
					"type": "number",
				},
			},
		},
	}),
)
```

## Messages

The interface `ChatMessage` represent a message sent or received by the model.
This interface has four implementations:
- [`SystemMessage`](#system-message)
- `UserMessage`
- `AssistantMessage`
- `ToolMessage`

### System message

With a simple string:
```go
mistral.NewSystemMessageFromString("You're a useful assistant.")
```

### User message


With a simple string:
```go
mistral.NewUserMessageFromString("Tell me a joke")
```

### Assistant message

Assistant messages are returned by the model. You unlikely need to create them manually.
But just in case, here is how to do it:

```go
mistral.NewAssistantMessageFromString("What do you call a fake noodle? An impasta!")
```

### Tool message

Tool messages are emitted after a tool was actually called. 
This type of message contains the tool's response.

Example with a simple string:
```go
mistral.NewToolMessage(
    "toll-name",
	"tool-call-id",
	mistral.ContentString("a string returned by the tool") // (1)
)
```

1. Could be either a content string or a list of chunks.
   See [message content](#message-content) below for more information.

## Message content

There are two types of content for a message:
- a simple string
- chunks

### Simple string

The easiest way to create a simple string message is to use the functions:
- `NewUserMessageFromString`
- `NewSystemMessageFromString`

You can also use 

## Links

- [Mistral's API documentation](https://docs.mistral.ai/api/endpoint/chat#operation-chat_completion_v1_chat_completions_post)