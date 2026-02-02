package mlflow_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomas-marquis/mistral-client/mlflow"
)

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
}
