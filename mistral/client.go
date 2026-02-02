package mistral

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/thomas-marquis/mistral-client/internal/shared"
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

	cacheConfig cacheConfig
	reqConfig   shared.RequestConfig
	baseHeaders map[string]string
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
		reqConfig: shared.RequestConfig{
			Verbose:         false,
			RetryMaxRetries: 3,
			RetryWaitMin:    200 * time.Millisecond,
			RetryWaitMax:    1 * time.Second,
			RetryStatusCodes: map[int]struct{}{
				http.StatusTooManyRequests:     {},
				http.StatusInternalServerError: {},
				http.StatusBadGateway:          {},
				http.StatusServiceUnavailable:  {},
				http.StatusGatewayTimeout:      {},
			},
		},

		cacheConfig: cacheConfig{cacheDir: DefaultCacheDir, enabled: false},
		baseHeaders: map[string]string{
			"Content-Type":  "application/json; charset=utf-8",
			"Authorization": "Bearer " + apiKey,
		},
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
		c.reqConfig.Verbose = verbose
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
		c.reqConfig.RetryMaxRetries = maxRetries
		c.reqConfig.RetryWaitMin = waitMin
		c.reqConfig.RetryWaitMax = waitMax
	}
}

// WithRetryStatusCodes overrides the list of HTTP status codes that should trigger a retry.
// If not specified, defaults are: 429, 500, 502, 503, 504.
func WithRetryStatusCodes(codes ...int) Option {
	return func(c *clientImpl) {
		if len(codes) == 0 {
			return
		}
		c.reqConfig.RetryStatusCodes = make(map[int]struct{})
		for _, code := range codes {
			c.reqConfig.RetryStatusCodes[code] = struct{}{}
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
