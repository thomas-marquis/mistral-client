package main

import (
	"context"
	"fmt"

	"github.com/thomas-marquis/mistral-client/mistral"
	"github.com/thomas-marquis/mistral-client/mlflow"
)

// This example needs a running-local MLflow server with a chat prompt named "python_dev_chat"

var (
	// System prompt:
	_ = `
You are a senior {{ .Occupation }}.

You must follow these guidelines:
- IMPORTANT! Don't make up, if you don't know just say "I don't know"
{{- range .Guidelines }}
- {{ if .IsImportant }}IMPORTANT! {{ end }}{{ .Text }}
{{- end }}

The current date is {{ now | date "2006-01-02" }}.
`

	// User prompt:
	_ = `
Answer the user's question:
{{ .Question }}

Additional information:
{{- range .AdditionalInformation }}
- {{ . }}
{{- end }}
`
)

func main() {
	pr, err := mlflow.NewPromptRegistry("http://localhost:5000")
	if err != nil {
		panic(err)
	}

	msg, err := mistral.MessagesFromRegisteredPrompt(context.Background(), pr,
		"python_dev_chat", "@go_template",
		map[string]any{
			"Occupation": "developer",
			"Guidelines": []map[string]any{
				{"IsImportant": true, "Text": "be nice"},
				{"IsImportant": false, "Text": "be respectful"},
			},
			"Question": "What is the meaning of life?",
			"AdditionalInformation": []string{
				"Life is like a box of chocolates, you never know what you're gonna get.",
				"The meaning of life is to find out what it is.",
			},
		},
		mlflow.WithGoTemplateRenderer())
	if err != nil {
		panic(err)
	}

	for _, m := range msg {
		fmt.Printf("Role: %s\nContent: %s\n\n", m.Role(), m.Content().String())
	}

	// Output:
	//  Role: system
	//  Content: You are a senior developer.
	//
	//	You must follow these guidelines:
	//	- IMPORTANT! Don't make up, if you don't know just say "I don't know"
	//	- IMPORTANT! be nice
	//	- be respectful
	//
	//	The current date is 2026-02-02.
	//
	//  Role: user
	//  Content: Answer the user's question:
	//	What is the meaning of life?
	//
	//	Additional information:
	//	- Life is like a box of chocolates, you never know what you're gonna get.
	//	- The meaning of life is to find out what it is.
}
