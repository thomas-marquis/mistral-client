# Constraint the output format

Use these options if you want to constrain the model's output format.
It can be useful, for example, if you want the model return a JSON with a specific structure.

- **`WithResponseTextFormat`**

It is the default option (you unlikely need to specify it).
Instruct the model that the expected response format is a text (without any specific structure, unless if it is specified in the prompt).

- **`WithResponseJsonObjectFormat`**

This option instructs the model that the expected response format is a JSON object but without any specific structure.
You can specify the expected structure in your prompt.

Then, use the method `Output` to parse the JSON object.

Example:

```go
req := mistral.NewChatCompletionRequest("mistral-small-latest",
	mistral.BaseMessage{
	    mistral.NewUserMessageFromString("Tell me a joke about cats. Return a JSON object with a 'joke field'."),
    },
	mistral.WithResponseJsonObjectFormat())
res, err := client.ChatCompletion(ctx, req)

msg := res.AssistantMessage()

var joke map[string]any // (1)
if err := msg.Output(&joke); err != nil {
    // handle error
}
```

1. Here, you can use a plain struct type instead

- **`WithResponseJsonSchema`**

This option instructs the model that the expected response format is a JSON object with a specific structure.
You can optionally specify a description for each field. That will help the model to figure out how to fill each field.

You can then use the method `Output` to parse the JSON object.

Example:

```go
req := mistral.NewChatCompletionRequest("mistral-small-latest", 
	messages, 
	mistral.WithResponseJsonSchema(mistral.NewObjectPropertyDefinition(
		map[string]mistral.PropertyDefinition{
			"joke":        {Type: "string", Description: "A joke."}, 
			"laugh_level": {
				Type: "integer", 
				Description: "How funny is the joke? In percentage.",
			    Default: 50},
		},
    )))

msg := res.AssistantMessage()

var joke map[string]any
if err := msg.Output(&joke); err != nil {
    // handle error
}
```