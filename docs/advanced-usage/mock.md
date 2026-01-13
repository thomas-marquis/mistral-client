# Testing: mock the client

Because `mistral.Client` is just an interface, it's possible to mock it.
Here, I'll give you an example with [gomock](https://github.com/uber-go/mock), but you can use any mocking framework.

First, install the library:

```bash
go install go.uber.org/mock/mockgen@latest
```

Gomock works with code generation. Start by creating an `gen.go` file at your project's root directory:

```bash
touch gen.go
```

Then, describe in this file what code you want to generate and where.
Here's an example:

```go
package main_test

import _ "go.uber.org/mock/gomock"

//go:generate mockgen -package mocks -destination mocks/client.go github.com/thomas-marquis/mistral-client/mistral Client
```

Run the command `go generate ./...` to generate the mock's code.
A `mocks/client.go` file will be created.

Eventually, you can use this mock in your tests:

```go
package my_test

import (
	"context"
	"reflect"
	"testing"
	
	"github.com/golang/mock/gomock"
	"github.com/thomas-marquis/mistral-client/mistral"
	"<your_project_package_id>/mocks"
)

var (
	ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
)

func TestSomething(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	mockClient := mocks.NewMockClient(mockCtrl)
	mockClient.EXPECT().
		ChatCompletion(
		    gomock.AssignableToTypeOf(ctxType),
            gomock.Eq(&mistralclient.ChatCompletionRequest{
                CompletionConfig: mistralclient.CompletionConfig{
                    ResponseFormat: &mistralclient.ResponseFormat{Type: "text"},
                },
                Model:    "mistral-small-latest",
                Messages: messages,
            }),
        ).
		Return(&mistral.ChatCompletionResponse{
			Choices: []mistralclient.ChatCompletionChoice{
				{Message: mistralclient.NewAssistantMessageFromString("Hello simple human being!")},
			},
        }, nil)

	// Use mockClient in your tests
}
```