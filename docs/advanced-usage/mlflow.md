# MLflow integrations

Prompts are important assets in any AI project. Their life cycle is at least as important as the code itself.
Furthermore, a bad prompt update can lead to unexpected behavior or even errors in your application.

That's why prompt registries exist.

They offer a convenient way to write and store your prompts. They also offer versioning and data governance features.
Although people who write and update prompts are not necessarily developers or technical people, a prompt registry is a great way to share your prompts with the rest of the team or across projects.
Thus, prompt registries are technically language-agnostic.
A prompt originally written for a Python project can be used in a Golang project.

`mistral-client` integrates with [MLflow Prompt Registry](https://mlflow.org/docs/latest/genai/prompt-registry/).
That means you can store your prompts on MLflow and use them in your Go code seamlessly.

!!! example "Check the following examples"

    - [Chat completion with MLflow integration](https://github.com/thomas-marquis/mistral-client/tree/main/examples/chat-completion-mlflow/main.go)
    - [Prompt rendering with Go templates](https://github.com/thomas-marquis/mistral-client/tree/main/examples/mlflow-prompts-go-template/main.go)

!!! note

    This feature has been tested with **MLflow 3.9.0**.

## Setup

These features are located in the `mistral-client`'s `mlflow` package.

Start by creating a `PromptRegistry`:

```go
import "github.com/thomas-marquis/mistral-client/mlflow"

// ...

registry, err := mlflow.NewPromptRegistry(mlflowUrl)
```

This function ensures the provided MLflow server is reachable and healthy. It returns an error if it is not the case.

Some options are available to customize the registry:

- `mlflow.WithHttpClient`: use a custom (or mock) HTTP client. Default is `*http.Client`.
- `mlflow.WithRetry`: set the number of retries and retry delay for HTTP requests.

## Create a prompt on MLflow

Please refer to the [MLflow documentation](https://mlflow.org/docs/latest/genai/prompt-registry/) to learn how to create a prompt.
This process may change depending on your MLflow version.

MLflow lets you define two types of prompts:

- **text**: a single text message
- **chat**: multiple messages

For each type, it is possible to use a templating format (see below).

## Build a message list from a stored prompt

### Text prompt

Let's start with a simple text prompt:

```
You're a professional {{occupation}}.

Please answer the following question:
{{question}}
```

This prompt has been stored on MLflow with the name `my_prompt_name` and 3 versions.
The version `2` is aliased as `production`.
By default, MLflow uses the alias `latest` to refer to the latest version (here `3`).

In our example, we need the production version of the prompt.

```go
messages, err := mistral.MessagesFromRegisteredPrompt(
	ctx, registry, "my_prompt_name", 
	"@production", // (1)
	map[string]any{ // (2)
	    "occupation": "web developer",
		"question": "Is still possible to find a job?",
    })
```

1. In the code, an alias always starts with `@`.
   But that's not the case in MLflow. Here the corresponding MLflow alias is just `production`.
2. Since the prompt contains templating (`{{...}}`), you need to provide the values to fill the placeholders.

This function will return a list of messages with formatted content.

### Chat prompt

Here, we define on MLflow the very same prompt as above, but using the chat format:

=== "System"

    ```
    You're a professional {{occupation}}.
    ```

=== "User"

    ```
    Please answer the following question:
    {{question}}
    ```

The code to load it is **exactly the same** as for the text prompt.
The library will automatically detect the chat format and return a list of messages.

Thus, you can change the prompt format without changing your code.

## Use templating in your prompt

As mentioned above, you can use templating in your prompt to fill placeholders with values.

`mistral-client` is able to render a template into a string thanks to a `PromptRenderer`.

Three implementations are available:

- **Raw renderer**: simply return the template as is.
- **Simple renderer**: use the `{{...}}` placeholders. It does not support nested placeholders.
- **Go template renderer**: use the [Go template syntax](https://pkg.go.dev/text/template). All the [Sprig](https://masterminds.github.io/sprig/) functions are available.

### Simple format

By default, MLflow handles a simple templating format, which uses `{{...}}` placeholders.
Thus, `mistral-client` uses the Simple format renderer by default, you don't need to specify anything.

### Go template format

What our prompt looks like on MLflow:

=== "System"

    ```
    You're a professional {{ .occupation | upper }}.
    ```

=== "User"

    ```
    Please answer the following question:
    {{ .question }}
    ```

Then, you need to specify the `mlflow.WithGoTemplateRenderer` option:

```go
messages, err := mistral.MessagesFromRegisteredPrompt(
    ctx, registry, "my_prompt_name",
    "@production",
    map[string]any{
        "occupation": "web developer",
        "question": "Is still possible to find a job?",
    },
    mlflow.WithGoTemplateRenderer()) // (1)
```

1. Specify which renderer to use as an option.

!!! tip

    Good documentation to learn how to use Go templates is the Helm documentation.
    Even though you're not familiar with Helm, here's a selection of some insightful pages from its documentation that may help you write templates:
   
    - [Ifs and loops](https://helm.sh/docs/chart_template_guide/control_structures#looping-with-the-range-action)
    - [Variable syntax](https://helm.sh/docs/chart_template_guide/variables)

### Other formats

If you want to use a different templating format, you can provide your own `PromptRenderer` implementation:

```go
type MyRenderer struct{}

func (r *MyRenderer) Render(raw string, params map[string]any) (string, error) {
	var result string
	// ...
    return result, nil
}
```

Then, use the `mlflow.WithPromptRenderer` option to specify your custom renderer:

```go
messages, err := mistral.MessagesFromRegisteredPrompt(
    ctx, registry, "my_prompt_name",
    "@production",
    map[string]any{
		// ...
    },
    mlflow.WithCustomPromptRenderer(&MyRenderer{})) // (1)
```