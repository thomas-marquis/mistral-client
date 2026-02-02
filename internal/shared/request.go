package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	Logger = log.New(os.Stdout, "mistral-client: ", log.LstdFlags|log.Lshortfile)
)

type RequestConfig struct {
	RetryMaxRetries  int
	RetryWaitMin     time.Duration
	RetryWaitMax     time.Duration
	RetryStatusCodes map[int]struct{}
	Verbose          bool
}

type ApiError struct {
	StatusCode int
	Content    map[string]any
}

func (e ApiError) Error() string {
	return fmt.Sprintf("API call error %d", e.StatusCode)
}

func SendRequest(
	ctx context.Context,
	httpClient *http.Client,
	method, url string,
	body []byte,
	header map[string]string,
	config RequestConfig,
) (*http.Response, time.Duration, error) {
	// attempt = 0 is the first try; we perform up to (1 + retryMaxRetries) attempts total.
	for attempt := 0; attempt <= config.RetryMaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			return nil, 0, fmt.Errorf("failed to create HTTP request: %w", err)
		}

		for hk, hv := range header {
			req.Header.Set(hk, hv)
		}

		t0 := time.Now()
		resp, err := httpClient.Do(req)
		latency := time.Since(t0)
		if err != nil {
			if attempt < config.RetryMaxRetries && isRetryableErr(err) {
				wait := nextBackoff(attempt, config)
				if config.Verbose {
					Logger.Printf("HTTP request error, retrying attempt %d/%d after %v: %v",
						attempt+1, config.RetryMaxRetries, wait, err)
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
			if attempt < config.RetryMaxRetries {
				if _, ok := config.RetryStatusCodes[resp.StatusCode]; ok {
					// Drain and close the body before retrying
					if _, err := io.Copy(io.Discard, resp.Body); err != nil {
						return nil, 0, fmt.Errorf("failed to drain response body: %w", err)
					}
					wait := nextBackoff(attempt, config)
					if config.Verbose {
						Logger.Printf("HTTP status %s, retrying attempt %d/%d after %v",
							resp.Status, attempt+1, config.RetryMaxRetries, wait)
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
				if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
					return nil, 0, ApiError{resp.StatusCode, nil}
				}
				return nil, 0, ApiError{resp.StatusCode, content}
			}

			errResponseBody, _ := io.ReadAll(resp.Body)
			return nil, 0, fmt.Errorf("HTTP request failed with status %s and body '%s'",
				resp.Status, string(errResponseBody))
		}

		return resp, latency, nil
	}

	return nil, 0, fmt.Errorf("exhausted retries without a successful response")
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

func nextBackoff(attempt int, config RequestConfig) time.Duration {
	if attempt <= 0 {
		return config.RetryWaitMin
	}
	wait := config.RetryWaitMin * time.Duration(1<<uint(attempt))
	if wait > config.RetryWaitMax {
		wait = config.RetryWaitMax
	}
	// Full jitter in [0, wait]
	jitter := time.Duration(rand.Int63n(int64(wait)))
	return jitter
}
