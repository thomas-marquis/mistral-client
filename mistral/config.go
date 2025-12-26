package mistral

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

type Config struct {
	ClientTimeout     time.Duration
	MistralAPIBaseURL string
	RateLimiter       *rate.Limiter
	ApiKey            string
	Verbose           bool

	// Retry configuration

	RetryMaxRetries  int
	RetryWaitMin     time.Duration
	RetryWaitMax     time.Duration
	RetryStatusCodes []int

	Transport http.RoundTripper
}

func NewConfig(opts ...Option) *Config {
	cfg := &Config{}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

type Option func(*Config)

func WithClientTimeout(timeout time.Duration) Option {
	return func(cfg *Config) {
		cfg.ClientTimeout = timeout
	}
}

func WithBaseApiUrl(baseURL string) Option {
	return func(cfg *Config) {
		cfg.MistralAPIBaseURL = strings.TrimSuffix(baseURL, "/")
	}
}

func WithRateLimiter(rateLimiter *rate.Limiter) Option {
	return func(cfg *Config) {
		cfg.RateLimiter = rateLimiter
	}
}

func WithVerbose(verbose bool) Option {
	return func(cfg *Config) {
		cfg.Verbose = verbose
	}
}

// WithRetry configures automatic retries for HTTP requests.
// maxRetries is the number of retries after the first attempt.
// waitMin and waitMax control the exponential backoff bounds (set to 0 for default).
func WithRetry(maxRetries int, waitMin, waitMax time.Duration) Option {
	if waitMin == 0 {
		waitMin = 200 * time.Millisecond
	}
	if waitMax == 0 {
		waitMax = 2 * time.Second
	}

	return func(cfg *Config) {
		cfg.RetryMaxRetries = maxRetries
		cfg.RetryWaitMin = waitMin
		cfg.RetryWaitMax = waitMax
	}
}

// WithRetryStatusCodes overrides the list of HTTP status codes that should trigger a retry.
// If not specified, defaults are: 429, 500, 502, 503, 504.
func WithRetryStatusCodes(codes ...int) Option {
	return func(cfg *Config) {
		cfg.RetryStatusCodes = append([]int(nil), codes...)
	}
}
