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
	"strings"
	"time"

	"github.com/thomas-marquis/mistral-client/mistral/internal/cache"
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

	// ChatCompletionStream calls the /v1/chat/completions endpoint with streaming enabled
	ChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (<-chan *CompletionChunk, error)

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

	cacheConfig cacheConfig
}

type Option func(impl *clientImpl)

// New create a new Client instance. Available options are:
//   - WithClientTimeout
//   - WithBaseApiUrl
//   - WithRateLimiter
//   - WithVerbose
//   - WithRetry
//   - WithRetryStatusCodes
//   - WithClientTransport
func New(apiKey string, opts ...Option) Client {
	c := &clientImpl{
		apiKey:  apiKey,
		baseURL: BaseApiUrl,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		verbose:          false,
		retryMaxRetries:  3,
		retryWaitMin:     200 * time.Millisecond,
		retryWaitMax:     1 * time.Second,
		retryStatusCodes: make(map[int]struct{}),

		cacheConfig: cacheConfig{cacheDir: DefaultCacheDir, enabled: false},
	}

	for _, code := range []int{
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	} {
		c.retryStatusCodes[code] = struct{}{}
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.cacheConfig.enabled {
		engine, err := cache.NewLocalFsEngine(c.cacheConfig.cacheDir) // TODO: implement other kind of engines later (s3, db...)
		if err != nil {
			logger.Fatalf("Failed to initialize local cache engine: %v", err)
		}

		return NewCached(c, engine)
	}

	return c
}

// NewCached decorates a client instance to cache responses with the given cache engine.
func NewCached(client Client, cacheEngine CacheEngine) Client {
	cc, err := newCachedClient(client, cacheEngine)
	if err != nil {
		logger.Fatalf("Failed to initialize local cache: %v", err)
	}
	return cc
}

func WithClientTimeout(timeout time.Duration) Option {
	return func(c *clientImpl) {
		c.httpClient.Timeout = timeout
	}
}

func WithBaseApiUrl(baseURL string) Option {
	return func(c *clientImpl) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

func WithRateLimiter(rateLimiter *rate.Limiter) Option {
	return func(c *clientImpl) {
		c.limiter = rateLimiter
	}
}

func WithVerbose(verbose bool) Option {
	return func(c *clientImpl) {
		c.verbose = verbose
	}
}

// WithRetry configures automatic retries for HTTP requests.
// maxRetries is the number of retries after the first attempt.
// waitMin and waitMax control the exponential backoff bounds (set to 0 for default).
// Accepted ranges:
//
//	0 < waitMin <= waitMax
func WithRetry(maxRetries int, waitMin, waitMax time.Duration) Option {
	if waitMin == 0 {
		waitMin = 200 * time.Millisecond
	}
	if waitMax == 0 {
		waitMax = 1 * time.Second
	}
	if waitMin > waitMax {
		waitMin, waitMax = waitMax, waitMin
	}

	return func(c *clientImpl) {
		c.retryMaxRetries = maxRetries
		c.retryWaitMin = waitMin
		c.retryWaitMax = waitMax
	}
}

// WithRetryStatusCodes overrides the list of HTTP status codes that should trigger a retry.
// If not specified, defaults are: 429, 500, 502, 503, 504.
func WithRetryStatusCodes(codes ...int) Option {
	return func(c *clientImpl) {
		if len(codes) == 0 {
			return
		}
		c.retryStatusCodes = make(map[int]struct{})
		for _, code := range codes {
			c.retryStatusCodes[code] = struct{}{}
		}
	}
}

// WithClientTransport overrides the underlying HTTP client transport.
func WithClientTransport(t http.RoundTripper) Option {
	return func(c *clientImpl) {
		c.httpClient.Transport = t
	}
}

// WithLocalCache enables caching of responses in the local file system.
// NewCached response will be stored in the DefaultCacheDir
func WithLocalCache() Option {
	return func(c *clientImpl) {
		c.cacheConfig.enabled = true
	}
}

// WithCacheDir enables local caching and sets the directory where cached responses will be stored.
func WithCacheDir(dir string) Option {
	return func(c *clientImpl) {
		c.cacheConfig.enabled = true
		c.cacheConfig.cacheDir = dir
	}
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
