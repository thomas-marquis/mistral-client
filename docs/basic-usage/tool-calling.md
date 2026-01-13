# Call tools

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
toolAdd := mistral.NewTool("add", "add two numbers", mistral.NewObjectPropertyDefinition(
    map[string]mistral.PropertyDefinition{
        "a": mistral.PropertyDefinition{Type: "number"},
        "b": mistral.PropertyDefinition{Type: "number"},
    })

toolMultiply := mistral.NewTool("multiply", "multiply two numbers", mistral.NewObjectPropertyDefinition(
    map[string]mistral.PropertyDefinition{
        "a": mistral.PropertyDefinition{Type: "number"},
        "b": mistral.PropertyDefinition{Type: "number"},
    })

req := mistral.NewChatCompletionRequest("mistral-small-latest",
	messages,
	mistral.WithTools([]mistral.Tool{toolAdd, toolMultiply}),
)

res, err := client.ChatCompletion(ctx, req)
```

## Handling the model response

When the model answer is received, you need to check the `ToolChoices` attribute in the message to know which tools were invoked (if any).

```go
msg := res.AssistantMessage()

if msg != nil && len(msg.ToolCalls) > 0 {
    for _, call := range msg.ToolCalls {
        fmt.Printf("Function %s called with arguments:\n%+v\n", call.Function.Name, call.Function.Arguments)
    }
}
```

Then, you have to call the actual functions implementations by yourself, according to the tool chosen by the model.

```go
func add(a, b int) int {
    return a + b
}

func multiply(a, b int) int {
    return a * b
}


func callTool(toolName string, args map[string]any) (any, error) {
	switch toolName {
	case "add":
		return add(args["a"].(int), args["b"].(int)), nil
	case "multiply":
		return multiply(args["a"].(int), args["b"].(int)), nil
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

func main() {
	...

	for _, call := range msg.ToolCalls {
        toolRes, err := callTool(call.Function.Name, call.Function.Arguments)
        if err != nil {
            // handle error
        }
		// cast and use toolRes
    }
}

```

## Create a tool response message

If you want to continue interacting with the model after a tool has been called, you need to create a `ToolMessage` with the tool call results and push it at the end of the message list.

```go
toolRes, err := callTool(call.Function.Name, call.Function.Arguments)

tooMsg := mistral.NewToolMessage(call.Function.Name, call.ID,
    mistral.ContentString(fmt.Sprintf("The result id: %d", toolRes.(int))))
```