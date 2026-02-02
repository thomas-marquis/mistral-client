package mlflow_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomas-marquis/mistral-client/mlflow"
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

func TestNewPromptRegistry(t *testing.T) {
	t.Run("should return an error if server is unreachable", func(t *testing.T) {
		// When
		_, err := mlflow.NewPromptRegistry("http://localhost:26829")

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("should return an error if server is not healthy", func(t *testing.T) {
		// Given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		// When
		_, err := mlflow.NewPromptRegistry(server.URL)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mlflow server is not healthy: 503")
	})
}

func setupMockMlflowServer(t *testing.T, respBody string, urlCapture *url.URL) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if urlCapture != nil {
			*urlCapture = *r.URL
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, respBody)
	}))
}

func TestPromptRegistry_Get(t *testing.T) {
	t.Run("should successfully get a prompt with latest version", func(t *testing.T) {
		// Given
		mockResponse := `{
		"model_version":	
			{
				"name": "python_dev_system",
				"version": "1",
				"creation_timestamp": 1769977763165,
				"last_updated_timestamp": 1769977763165,
				"current_stage": "None",
				"description": "init",
				"source": "dummy-source",
				"run_id": "",
				"status": "READY",
				"tags": [
					{
						"key": "mlflow.prompt.is_prompt",
						"value": "true"
					},
					{
						"key": "mlflow.prompt.text",
						"value": "You are a senior python developer.\nYou're experienced into developing complex and modular application following the clean architecture principles.\nTake a deep breath and proceed step by step:\n- first think about the folder structure\n- then implement each file"
					},
					{
						"key": "_mlflow_prompt_type",
						"value": "text"
					}
				],
				"run_link": ""
			}
		}`
		var actualURL url.URL
		server := setupMockMlflowServer(t, mockResponse, &actualURL)
		defer server.Close()

		registry, err := mlflow.NewPromptRegistry(server.URL)
		require.NoError(t, err)

		// When
		prompt, err := registry.Get(context.TODO(), "python_dev_system", mlflow.VersionLatest)
		res, ok := prompt.(*mlflow.PromptText)
		require.True(t, ok)

		// Then
		require.NoError(t, err)
		assert.NotNil(t, prompt)
		assert.Equal(t, "python_dev_system", prompt.Name())
		assert.Equal(t, mlflow.Version("1"), prompt.Version())
		assert.Contains(t, res.RawTextTemplate(), "You are a senior python developer.")
		assert.Contains(t, res.RawTextTemplate(), "clean architecture principles")
		assert.Equal(t, mlflow.PromptTypeText, prompt.Type())

		assert.Equal(t, "/api/2.0/mlflow/registered-models/alias", actualURL.Path)
		assert.Equal(t, "python_dev_system", actualURL.Query().Get("name"))
		assert.Equal(t, "latest", actualURL.Query().Get("alias"))
	})

	t.Run("should successfully get a prompt with specific version", func(t *testing.T) {
		// Given
		mockResponse := `{
			"model_version": {
				"name": "python_dev_system",
				"version": "2",
				"creation_timestamp": 1770017643354,
				"last_updated_timestamp": 1770017643354,
				"current_stage": "None",
				"description": "add refactoring step",
				"source": "dummy-source",
				"run_id": "",
				"status": "READY",
				"tags": [
					{
						"key": "mlflow.prompt.is_prompt",
						"value": "true"
					},
					{
						"key": "mlflow.prompt.text",
						"value": "You are a senior python developer.\nYou're experienced into developing complex and modular application following the clean architecture principles.\nTake a deep breath and proceed step by step:\n- first think about the folder structure\n- then implement each file\n- eventually, review your code and refactor it"
					},
					{
						"key": "_mlflow_prompt_type",
						"value": "text"
					}
				],
				"run_link": ""
			}
		}`
		var actualURL url.URL
		server := setupMockMlflowServer(t, mockResponse, &actualURL)
		defer server.Close()

		registry, err := mlflow.NewPromptRegistry(server.URL)
		require.NoError(t, err)

		// When
		prompt, err := registry.Get(context.TODO(), "python_dev_system", "2")
		res, ok := prompt.(*mlflow.PromptText)
		require.True(t, ok)

		// Then
		require.NoError(t, err)
		assert.Equal(t, mlflow.Version("2"), prompt.Version())
		assert.Equal(t, "python_dev_system", prompt.Name())
		assert.Contains(t, res.RawTextTemplate(), "review your code and refactor it")
		assert.Equal(t, mlflow.PromptTypeText, prompt.Type())

		assert.Equal(t, "/api/2.0/mlflow/model-versions/get", actualURL.Path)
		assert.Equal(t, "python_dev_system", actualURL.Query().Get("name"))
		assert.Equal(t, "2", actualURL.Query().Get("version"))
	})

	t.Run("should successfully get a prompt with with alias", func(t *testing.T) {
		// Given
		mockResponse := `{
			"model_version": {
				"name": "python_dev_system",
				"version": "2",
				"creation_timestamp": 1770017643354,
				"last_updated_timestamp": 1770017643354,
				"current_stage": "None",
				"description": "add refactoring step",
				"source": "dummy-source",
				"run_id": "",
				"status": "READY",
				"tags": [
					{
						"key": "mlflow.prompt.is_prompt",
						"value": "true"
					},
					{
						"key": "mlflow.prompt.text",
						"value": "You are a senior python developer.\nYou're experienced into developing complex and modular application following the clean architecture principles.\nTake a deep breath and proceed step by step:\n- first think about the folder structure\n- then implement each file\n- eventually, review your code and refactor it"
					},
					{
						"key": "_mlflow_prompt_type",
						"value": "text"
					}
				],
				"run_link": "",
				"aliases": [
					"production"
				]
			}
		}`
		var actualURL url.URL
		server := setupMockMlflowServer(t, mockResponse, &actualURL)
		defer server.Close()

		registry, err := mlflow.NewPromptRegistry(server.URL)
		require.NoError(t, err)

		// When
		prompt, err := registry.Get(context.TODO(), "python_dev_system", "@production")
		res, ok := prompt.(*mlflow.PromptText)
		require.True(t, ok)

		// Then
		require.NoError(t, err)
		assert.NotNil(t, prompt)
		assert.Equal(t, "python_dev_system", prompt.Name())
		assert.Equal(t, mlflow.Version("2"), prompt.Version())
		assert.Contains(t, res.RawTextTemplate(), "You are a senior python developer.")
		assert.Contains(t, res.RawTextTemplate(), "review your code and refactor it")
		assert.Equal(t, mlflow.PromptTypeText, prompt.Type())

		assert.Equal(t, "/api/2.0/mlflow/registered-models/alias", actualURL.Path)
		assert.Equal(t, "python_dev_system", actualURL.Query().Get("name"))
		assert.Equal(t, "production", actualURL.Query().Get("alias"))
	})

	t.Run("should get a chat prompt type", func(t *testing.T) {
		// Given
		mockResponse := `{
			"model_version": {
				"name": "python_dev_chat",
				"version": "1",
				"creation_timestamp": 1770018239526,
				"last_updated_timestamp": 1770018239526,
				"current_stage": "None",
				"description": "init",
				"source": "dummy-source",
				"run_id": "",
				"status": "READY",
				"tags": [
					{
						"key": "mlflow.prompt.is_prompt",
						"value": "true"
					},
					{
						"key": "mlflow.prompt.text",
						"value": "[{\"role\":\"system\",\"content\":\"You are a senior python developer.\\nYou're experienced into developing complex and modular application following the clean architecture principles.\\nTake a deep breath and proceed step by step:\\n- first think about the folder structure\\n- then implement each file\"},{\"role\":\"user\",\"content\":\"Write a simple and nice tkinter application in a single file named main.py.\\nSpecifications:\\n{{specifications}}\"}]"
					},
					{
						"key": "_mlflow_prompt_type",
						"value": "chat"
					}
				],
				"run_link": ""
			}
		}`
		var actualURL url.URL
		server := setupMockMlflowServer(t, mockResponse, &actualURL)
		defer server.Close()

		registry, err := mlflow.NewPromptRegistry(server.URL)
		require.NoError(t, err)

		// When
		prompt, err := registry.Get(context.TODO(), "python_dev_chat", "1")

		// Then
		require.NoError(t, err)
		assert.Equal(t, mlflow.Version("1"), prompt.Version())
		assert.Equal(t, "python_dev_chat", prompt.Name())
		assert.Equal(t, mlflow.PromptTypeChat, prompt.Type())
		res, ok := prompt.(*mlflow.PromptChat)
		assert.True(t, ok)

		msgs := res.RawMessagesTemplate()
		assert.Len(t, msgs, 2)

		assert.Equal(t, mlflow.PromptRoleSystem, msgs[0].Role)
		assert.Contains(t, msgs[0].Content, "following the clean architecture principles")

		assert.Equal(t, mlflow.PromptRoleUser, msgs[1].Role)
		assert.Contains(t, msgs[1].Content, "Write a simple and nice tkinter application in a single file named main.py")

		assert.Equal(t, "/api/2.0/mlflow/model-versions/get", actualURL.Path)
		assert.Equal(t, "python_dev_chat", actualURL.Query().Get("name"))
		assert.Equal(t, "1", actualURL.Query().Get("version"))
	})

	t.Run("should return error if version not found", func(t *testing.T) {
		// Given
		mockResponse := `{
			"error_code": "RESOURCE_DOES_NOT_EXIST",
			"message": "Model Version (name=python_dev_system, version=3) not found"
		}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintln(w, mockResponse)
		}))
		defer server.Close()

		registry, err := mlflow.NewPromptRegistry(server.URL)
		require.NoError(t, err)

		// When
		prompt, err := registry.Get(context.TODO(), "python_dev_system", "3")

		// Then
		assert.Error(t, err)
		assert.Nil(t, prompt)

		var apiErr *mlflow.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Contains(t, apiErr.Error(), "Model Version (name=python_dev_system, version=3) not found")
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, "RESOURCE_DOES_NOT_EXIST", apiErr.ErrorCode)
	})

	t.Run("should return APIError when prompt name doesn't exist", func(t *testing.T) {
		// Given
		mockResponse := `{
			"error_code": "RESOURCE_DOES_NOT_EXIST",
			"message": "Model Version (name=no_exists, version=2) not found"
		}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintln(w, mockResponse)
		}))
		defer server.Close()

		registry, err := mlflow.NewPromptRegistry(server.URL)
		require.NoError(t, err)

		// When
		prompt, err := registry.Get(context.TODO(), "no_exists", "2")

		// Then
		assert.Error(t, err)
		assert.Nil(t, prompt)

		var apiErr *mlflow.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, "RESOURCE_DOES_NOT_EXIST", apiErr.ErrorCode)
		assert.Contains(t, apiErr.Message, "Model Version (name=no_exists, version=2) not found")
	})

	t.Run("Should retry on 5xx then succeed", func(t *testing.T) {
		// Given
		var attempts int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			atomic.AddInt32(&attempts, 1)
			if atomic.LoadInt32(&attempts) <= 2 {
				http.Error(w, `{"error_code":"INTERNAL_ERROR","message":"temporary"}`, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"model_version": {
					"name": "retry_model",
					"version": "1",
					"tags": [{"key": "mlflow.prompt.text", "value": "success after retry"}, {"key": "_mlflow_prompt_type", "value": "text"}]
				}
			}`))
		}))
		defer srv.Close()

		registry, err := mlflow.NewPromptRegistry(srv.URL,
			mlflow.WithRetry(3, 1*time.Millisecond, 5*time.Millisecond),
		)
		require.NoError(t, err)

		// When
		prompt, err := registry.Get(context.Background(), "retry_model", "1")

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, prompt)
		assert.Equal(t, "success after retry", prompt.(*mlflow.PromptText).RawTextTemplate())
		assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	})

	t.Run("Should not retry on 404 and fail immediately", func(t *testing.T) {
		// Given
		var attempts int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			atomic.AddInt32(&attempts, 1)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{
				"error_code": "RESOURCE_DOES_NOT_EXIST",
				"message": "not found"
			}`))
		}))
		defer srv.Close()

		registry, err := mlflow.NewPromptRegistry(srv.URL,
			mlflow.WithRetry(5, 1*time.Millisecond, 2*time.Millisecond),
		)
		require.NoError(t, err)

		// When
		_, err = registry.Get(context.Background(), "not_found_model", "1")

		// Then
		assert.Error(t, err)
		var apiErr *mlflow.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	})

	t.Run("Should retry on timeout error then succeed", func(t *testing.T) {
		// Given
		successJSON := []byte(`{
			"model_version": {
				"name": "timeout_model",
				"version": "1",
				"tags": [{"key": "mlflow.prompt.text", "value": "OK after timeout"}, {"key": "_mlflow_prompt_type", "value": "text"}]
			}
		}`)

		// We need to bypass health check or make it succeed
		healthSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer healthSrv.Close()

		registry, err := mlflow.NewPromptRegistry(healthSrv.URL,
			mlflow.WithRetry(3, 1*time.Millisecond, 5*time.Millisecond),
			mlflow.WithHttpClient(&http.Client{
				Transport: &flakyRoundTripper{
					failuresLeft: 1,
					successBody:  successJSON,
				},
			}),
		)
		require.NoError(t, err)

		// When
		prompt, err := registry.Get(context.Background(), "timeout_model", "1")

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, prompt)
		assert.Equal(t, "OK after timeout", prompt.(*mlflow.PromptText).RawTextTemplate())
	})

	t.Run("Should fail when max retries reached", func(t *testing.T) {
		// Given
		var attempts int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			atomic.AddInt32(&attempts, 1)
			http.Error(w, `{"error_code":"UNAVAILABLE","message":"unavailable"}`, http.StatusServiceUnavailable)
		}))
		defer srv.Close()

		registry, err := mlflow.NewPromptRegistry(srv.URL,
			mlflow.WithRetry(2, 1*time.Millisecond, 2*time.Millisecond),
		)
		require.NoError(t, err)

		// When
		_, err = registry.Get(context.Background(), "fail_model", "1")

		// Then
		assert.Error(t, err)
		assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	})
}
