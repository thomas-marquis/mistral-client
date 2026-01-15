package mistral_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// timeoutNetError implements net.Error with Timeout() = true
type timeoutNetError struct{}

func (timeoutNetError) Error() string   { return "timeout" }
func (timeoutNetError) Timeout() bool   { return true }
func (timeoutNetError) Temporary() bool { return true } // for legacy checks

// flakyRoundTripper fails with a timeout once, then returns a successful response.
type flakyRoundTripper struct {
	failuresLeft int32
	successBody  []byte
}

func (f *flakyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if atomic.AddInt32(&f.failuresLeft, -1) >= 0 {
		return nil, timeoutNetError{}
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(f.successBody)),
		Request:    req,
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}

// makeMockServerWithCapture creates an HTTP test server that returns a fixed JSON response
// and also captures the JSON request body. The captured JSON is pretty-printed to make
// assertions easy to read in tests.
func makeMockServerWithCapture(t *testing.T, method, path, jsonResponse string, responseCode int, capturedBody *string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method && r.URL.Path == path {
			// Capture and pretty-print the incoming JSON body for assertions
			if r.Body != nil {
				defer r.Body.Close() //nolint:errcheck
				raw, _ := io.ReadAll(r.Body)
				var anyJSON any
				if len(bytes.TrimSpace(raw)) > 0 && json.Unmarshal(raw, &anyJSON) == nil {
					pretty, _ := json.MarshalIndent(anyJSON, "", "  ")
					*capturedBody = string(pretty)
				} else {
					*capturedBody = string(raw)
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(responseCode)
			_, _ = w.Write([]byte(jsonResponse))
			return
		}
		http.NotFound(w, r)
	}))
}

func makeMockSseServerWithCapture(t *testing.T, method, path string, messages []string, responseCode int, capturedBody *string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(messages) == 0 {
			http.Error(w, "no messages provided", http.StatusInternalServerError)
			return
		}
		if strings.TrimSpace(messages[len(messages)-1]) != "data: [DONE]" {
			http.Error(w, "no special done message found", http.StatusInternalServerError)
			return
		}

		if r.Method == method && r.URL.Path == path {
			if r.Body != nil {
				defer r.Body.Close() //nolint:errcheck
				raw, _ := io.ReadAll(r.Body)
				var anyJSON any
				if len(bytes.TrimSpace(raw)) > 0 && json.Unmarshal(raw, &anyJSON) == nil {
					pretty, _ := json.MarshalIndent(anyJSON, "", "  ")
					*capturedBody = string(pretty)
				} else {
					*capturedBody = string(raw)
				}
			}

			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(responseCode)

			var wg sync.WaitGroup
			messageChan := make(chan string)
			clientGone := r.Context().Done()

			go func() {
				for {
					select {
					case <-clientGone:
						log.Println("Client disconnected")
						return
					case msg := <-messageChan:
						if _, err := fmt.Fprintf(w, "%s\n", msg); err != nil {
							log.Println("Failed to write message:", err)
							wg.Done()
							return
						}
						if f, ok := w.(http.Flusher); ok {
							f.Flush()
						}
						if strings.TrimSpace(msg) == "data: [DONE]" {
							wg.Done()
							close(messageChan)
							return
						}
						if _, err := fmt.Fprint(w, "\n\n"); err != nil {
							log.Println("Failed to write message:", err)
							wg.Done()
							return
						}
						if f, ok := w.(http.Flusher); ok {
							f.Flush()
						}
						wg.Done()
					}
				}
			}()

			ticker := time.NewTicker(3 * time.Millisecond)
			defer ticker.Stop()

			for _, data := range messages {
				wg.Add(1)
				select {
				case <-clientGone:
					log.Println("Client disconnected")
					return
				case <-ticker.C:
					messageChan <- data
				}
			}

			wg.Wait()

			return
		}
		http.NotFound(w, r)
	}))
}
