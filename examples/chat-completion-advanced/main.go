package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/thomas-marquis/mistral-client/mistral"
	"golang.org/x/time/rate"
)

type fakeRoundTripper struct {
	Body []byte
}

func (t *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(t.Body)),
		Request:    req,
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}

func main() {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		panic("Please set MISTRAL_API_KEY environment variable")
	}
	client := mistral.New(apiKey,
		mistral.WithClientTimeout(60*time.Second),
		mistral.WithRateLimiter(rate.NewLimiter(rate.Every(1*time.Second), 50)),
		mistral.WithBaseApiUrl(mistral.BaseApiUrl),
		mistral.WithVerbose(true),
		mistral.WithRetry(4, 1*time.Second, 3*time.Second),
		mistral.WithRetryStatusCodes(429, 500, 502, 503, 504),
		mistral.WithClientTransport(&fakeRoundTripper{Body: []byte(`{
			"choices": [
				{ "message": { "role": "assistant", "content": "Hello from fake" } }
			]
		}`)}), // This option may be used for testing purposes
	)

	systemPrompt := `You are a useful assistant.`
	userPrompt := `Because the transport is mocked, the actual API wont be called...`

	req := mistral.NewChatCompletionRequest(
		"mistral-small-latest",
		[]mistral.ChatMessage{
			mistral.NewSystemMessageFromString(systemPrompt),
			mistral.NewUserMessageFromString(userPrompt),
		})
	req.MaxTokens = 128_000
	res, err := client.ChatCompletion(context.Background(), req)
	if err != nil {
		panic(err)
	}

	msg := res.AssistantMessage()
	if msg != nil {
		fmt.Println(msg.MessageContent)
	} else {
		panic("No assistant message found")
	}

	fmt.Printf("Latency: %fs\n", res.Latency.Seconds())
	fmt.Printf("Input tokens: %d; Completion tokens: %d; Total tokens: %d\n",
		res.Usage.PromptTokens, res.Usage.CompletionTokens, res.Usage.TotalTokens)
}
