# Tool calling

Tool calling enables the possibility to call external tools from the model.

The model may (or may not) decide to call one or more tools among the ones specified.

The models won't call the tools directly (because the tools are external to the model), but will tell you in the response:

- with tool(s) invoked
- with their input and output

## Specify tools

Specify a list of tools to use with the `WithTools` option.

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

## Handling tool responses

When the model answer is received, you need to check the `ToolChoices` attribute in the message in order to know which tools were invoked (if any.

```go

```