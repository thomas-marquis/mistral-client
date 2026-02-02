package mlflow_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomas-marquis/mistral-client/mlflow"
)

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
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/2.0/mlflow/registered-models/alias", r.URL.Path)
			assert.Equal(t, "python_dev_system", r.URL.Query().Get("name"))
			assert.Equal(t, "latest", r.URL.Query().Get("alias"))
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintln(w, mockResponse)
		}))
		defer server.Close()

		registry := mlflow.NewPromptRegistry(server.URL)

		// When
		prompt, err := registry.Get("python_dev_system", mlflow.VersionLatest)

		// Then
		require.NoError(t, err)
		assert.NotNil(t, prompt)
		assert.Equal(t, "python_dev_system", prompt.Name())
		assert.Equal(t, mlflow.Version("1"), prompt.Version())
		assert.Contains(t, prompt.String(), "You are a senior python developer.")
		assert.Contains(t, prompt.String(), "clean architecture principles")
		assert.Equal(t, mlflow.PromptTypeText, prompt.Type())
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
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/2.0/mlflow/model-versions/get", r.URL.Path)
			assert.Equal(t, "python_dev_system", r.URL.Query().Get("name"))
			assert.Equal(t, "2", r.URL.Query().Get("version"))
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintln(w, mockResponse)
		}))
		defer server.Close()

		registry := mlflow.NewPromptRegistry(server.URL)

		// When
		prompt, err := registry.Get("python_dev_system", "2")

		// Then
		require.NoError(t, err)
		assert.Equal(t, mlflow.Version("2"), prompt.Version())
		assert.Equal(t, "python_dev_system", prompt.Name())
		assert.Contains(t, prompt.String(), "review your code and refactor it")
		assert.Equal(t, mlflow.PromptTypeText, prompt.Type())
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
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/2.0/mlflow/model-versions/get", r.URL.Path)
			assert.Equal(t, "python_dev_chat", r.URL.Query().Get("name"))
			assert.Equal(t, "1", r.URL.Query().Get("version"))
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintln(w, mockResponse)
		}))
		defer server.Close()

		registry := mlflow.NewPromptRegistry(server.URL)

		// When
		prompt, err := registry.Get("python_dev_chat", "1")

		// Then
		require.NoError(t, err)
		assert.Equal(t, mlflow.Version("1"), prompt.Version())
		assert.Equal(t, "python_dev_chat", prompt.Name())
		assert.Contains(t, prompt.String(), "review your code and refactor it")
		assert.Equal(t, mlflow.PromptTypeChat, prompt.Type())
	})

	t.Run("should return error if version not found", func(t *testing.T) {
		// Given
		mockResponse := `{
			"error_code": "RESOURCE_DOES_NOT_EXIST",
			"message": "Model Version (name=python_dev_system, version=3) not found"
		}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintln(w, mockResponse)
		}))
		defer server.Close()

		registry := mlflow.NewPromptRegistry(server.URL)

		// When
		prompt, err := registry.Get("python_dev_system", "3")

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
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintln(w, mockResponse)
		}))
		defer server.Close()

		registry := mlflow.NewPromptRegistry(server.URL)

		// When
		prompt, err := registry.Get("no_exists", "2")

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
