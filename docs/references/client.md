# Client

## Creating a client

Create a new client with the `mistral.New` function.

**Arguments:** 

- apiKey: `string`
- `...mistral.Option`

`mistral.Client` is an interface. An implementation is provided by the library, but feel free to write your own.

## Available options

### `WithClientTimeout`

**Arguments:** `time.Duration`

**Default value:** 30 seconds

### `WithBaseApiUrl`

Set the base URL for the API. Useful for testing.

**Arguments:** `string`

**Default value:** `https://api.github.com`

### `WithRateLimiter`

Configure a rate limiter with the package [`golang.org/x/time/rate`](https://pkg.go.dev/golang.org/x/time/rate)

**Arguments:** `rate.Limiter`

**Default value:** `nil` (no rate limiting applied)

### `WithVerbose`

Enable verbose mode.

**Arguments:** `bool`

**Default value:** `false`

### `WithRetry`

Customize the retry strategy.

**Arguments:**

- maxRetries: `int`. Max number of retries before giving up. Defaults to 3.
- waitMin: `time.Duration`. Minimum wait time between retries. Defaults to 200 milliseconds.
- waitMax: `time.Duration`. Maximum wait time between retries. Defaults to 1 second.

Constraints: `0 < waitMin <= waitMax`

### `WithRetryStatusCodes`

Customize which HTTP status codes should be retried.

**Arguments:** `...int`

**Default value:** `429, 500, 502, 503, 504`

Setting up this option will override the default status codes.

### `WithClientTransport`

Customize the HTTP transport implementation to use. Useful for testing.

**Arguments:** `http.RoundTripper`

**Default value:** `nil` (default transport from `http.Client`)
