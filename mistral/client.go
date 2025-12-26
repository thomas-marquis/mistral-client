package mistral

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	BaseApiUrl     = "https://api.mistral.ai"
	defaultTimeout = 30 * time.Second
)

type Client interface {
	// Embeddings calls the /v1/embeddings endpoint
	Embeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// ChatCompletion calls the /v1/chat/completions endpoint
	ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error)

	// ListModels lists all models available to the user.
	ListModels(ctx context.Context) ([]*BaseModelCard, error)

	// SearchModels searches for models that match the specified capabilities.
	// The returned models match at least all the specified capabilities.
	SearchModels(ctx context.Context, capabilities *ModelCapabilities) ([]*BaseModelCard, error)

	// GetModel returns the model card corresponding to the specified ID or an error if it does not exist.
	GetModel(ctx context.Context, modelId string) (*BaseModelCard, error)
}

type clientImpl struct {
	apiKey  string
	baseURL string

	limiter    *rate.Limiter
	httpClient *http.Client
	verbose    bool

	retryMaxRetries  int
	retryWaitMin     time.Duration
	retryWaitMax     time.Duration
	retryStatusCodes map[int]struct{}
}

// New create a new Client instance. Available options are:
//   - WithClientTimeout
//   - WithBaseApiUrl
//   - WithRateLimiter
//   - WithVerbose
//   - WithRetry
//   - WithRetryStatusCodes
func New(apiKey string, opts ...Option) Client {
	return NewWithConfig(apiKey, NewConfig(opts...))
}

func NewWithConfig(apiKey string, cfg *Config) Client {
	c := &clientImpl{
		apiKey:  apiKey,
		baseURL: BaseApiUrl,
		httpClient: &http.Client{
			Timeout:   defaultTimeout,
			Transport: cfg.Transport,
		},
		verbose:          cfg.Verbose,
		retryMaxRetries:  cfg.RetryMaxRetries,
		retryWaitMin:     cfg.RetryWaitMin,
		retryWaitMax:     cfg.RetryWaitMax,
		retryStatusCodes: make(map[int]struct{}),
	}

	if cfg.MistralAPIBaseURL != "" {
		c.baseURL = cfg.MistralAPIBaseURL
	}
	if cfg.RateLimiter != nil {
		c.limiter = cfg.RateLimiter
	}
	if cfg.ClientTimeout > 0 {
		c.httpClient.Timeout = cfg.ClientTimeout
	}
	if cfg.ApiKey != "" {
		c.apiKey = cfg.ApiKey
	}

	// status codes to retry: configured or sensible defaults
	if len(cfg.RetryStatusCodes) > 0 {
		for _, code := range cfg.RetryStatusCodes {
			c.retryStatusCodes[code] = struct{}{}
		}
	} else {
		for _, code := range []int{
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		} {
			c.retryStatusCodes[code] = struct{}{}
		}
	}

	return c
}

// isRetryableErr returns true if the error is retryable.
//
// Retriable errors:
//   - [net.Error] with Temporary() == true
//   - [context.DeadlineExceeded]
//   - unexpected EOFs ([io.EOF]) and similar transient I/O issues.
//
// Errors that are not retriable:
//   - [context.Canceled]
//   - any other errors
func isRetryableErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
		// Temporary is deprecated but still implemented by some errors.
		if te, ok := any(netErr).(interface{ Temporary() bool }); ok && te.Temporary() {
			return true
		}
	}

	return errors.Is(err, io.EOF)
}

func (c *clientImpl) nextBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return c.retryWaitMin
	}
	wait := c.retryWaitMin * time.Duration(1<<uint(attempt))
	if wait > c.retryWaitMax {
		wait = c.retryWaitMax
	}
	// Full jitter in [0, wait]
	jitter := time.Duration(rand.Int63n(int64(wait)))
	return jitter
}

func unmarshallBody(resp *http.Response, v interface{}) error {
	err := json.NewDecoder(resp.Body).Decode(v)
	return err
}

func (c *clientImpl) sendRequest(ctx context.Context, method, url string, body []byte) (*http.Response, time.Duration, error) {
	// attempt = 0 is the first try; we perform up to (1 + retryMaxRetries) attempts total.
	for attempt := 0; attempt <= c.retryMaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			return nil, 0, fmt.Errorf("failed to create HTTP request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Accept", "application/json; charset=utf-8")
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		t0 := time.Now()
		resp, err := c.httpClient.Do(req)
		latency := time.Since(t0)
		if err != nil {
			if attempt < c.retryMaxRetries && isRetryableErr(err) {
				wait := c.nextBackoff(attempt)
				if c.verbose {
					logger.Printf("HTTP request error, retrying attempt %d/%d after %v: %v",
						attempt+1, c.retryMaxRetries, wait, err)
				}
				select {
				case <-time.After(wait):
					continue
				case <-ctx.Done():
					return nil, 0, ctx.Err()
				}
			}
			return nil, 0, fmt.Errorf("failed to make HTTP request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			if attempt < c.retryMaxRetries {
				if _, ok := c.retryStatusCodes[resp.StatusCode]; ok {
					// Drain and close the body before retrying
					if _, err := io.Copy(io.Discard, resp.Body); err != nil {
						return nil, 0, fmt.Errorf("failed to drain response body: %w", err)
					}
					wait := c.nextBackoff(attempt)
					if c.verbose {
						logger.Printf("HTTP status %s, retrying attempt %d/%d after %v",
							resp.Status, attempt+1, c.retryMaxRetries, wait)
					}
					select {
					case <-time.After(wait):
						continue
					case <-ctx.Done():
						return nil, 0, ctx.Err()
					}
				}
			}

			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				var content map[string]any
				if err := unmarshallBody(resp, &content); err != nil {
					return nil, 0, NewApiError(resp.StatusCode, nil)
				}
				return nil, 0, NewApiError(resp.StatusCode, content)
			}

			errResponseBody, _ := io.ReadAll(resp.Body)
			return nil, 0, fmt.Errorf("HTTP request failed with status %s and body '%s'",
				resp.Status, string(errResponseBody))
		}

		return resp, latency, nil
	}

	return nil, 0, fmt.Errorf("exhausted retries without a successful response")
}
