package mistral_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/mistral-client/mistral"
)

const (
	listModelJsonResp = `
{
	"object": "list",
	"data": [
		{
			"id": "mistral-medium-latest",
			"object": "model",
			"created": 1766520693,
			"owned_by": "mistralai",
			"capabilities": {
				"completion_chat": true,
				"function_calling": true,
				"completion_fim": false,
				"fine_tuning": true,
				"vision": true,
				"ocr": false,
				"classification": false,
				"moderation": false,
				"audio": false
			},
			"name": "mistral-medium-2508",
			"description": "Update on Mistral Medium 3 with improved capabilities.",
			"max_context_length": 131072,
			"aliases": [
				"mistral-medium-2508",
				"mistral-medium"
			],
			"deprecation": null,
			"deprecation_replacement_model": null,
			"default_model_temperature": 0.3,
			"type": "base"
		},
		{
			"id": "open-mistral-nemo",
			"object": "model",
			"created": 1766520693,
			"owned_by": "mistralai",
			"capabilities": {
				"completion_chat": true,
				"function_calling": true,
				"completion_fim": false,
				"fine_tuning": true,
				"vision": false,
				"ocr": false,
				"classification": false,
				"moderation": false,
				"audio": false
			},
			"name": "open-mistral-nemo",
			"description": "Our best multilingual open source model released July 2024.",
			"max_context_length": 131072,
			"aliases": [
				"open-mistral-nemo-2407",
				"mistral-tiny-2407",
				"mistral-tiny-latest"
			],
			"deprecation": null,
			"deprecation_replacement_model": null,
			"default_model_temperature": 0.3,
			"type": "base"
		},
		{
			"id": "devstral-small-2507",
			"object": "model",
			"created": 1766520693,
			"owned_by": "mistralai",
			"capabilities": {
				"completion_chat": true,
				"function_calling": true,
				"completion_fim": false,
				"fine_tuning": false,
				"vision": false,
				"ocr": false,
				"classification": false,
				"moderation": false,
				"audio": false
			},
			"name": "devstral-small-2507",
			"description": "Our small open-source code-agentic model.",
			"max_context_length": 131072,
			"aliases": [],
			"deprecation": null,
			"deprecation_replacement_model": null,
			"default_model_temperature": 0.0,
			"type": "base"
		},
		{
			"id": "voxtral-mini-2507",
			"object": "model",
			"created": 1766520693,
			"owned_by": "mistralai",
			"capabilities": {
				"completion_chat": true,
				"function_calling": false,
				"completion_fim": false,
				"fine_tuning": false,
				"vision": false,
				"ocr": false,
				"classification": false,
				"moderation": false,
				"audio": true
			},
			"name": "voxtral-mini-2507",
			"description": "A mini audio understanding model released in July 2025",
			"max_context_length": 32768,
			"aliases": [
				"voxtral-mini-latest"
			],
			"deprecation": null,
			"deprecation_replacement_model": null,
			"default_model_temperature": 0.2,
			"type": "base"
		},
		{
			"id": "mistral-ocr-2512",
			"object": "model",
			"created": 1766520693,
			"owned_by": "mistralai",
			"capabilities": {
				"completion_chat": false,
				"function_calling": true,
				"completion_fim": false,
				"fine_tuning": false,
				"vision": true,
				"ocr": true,
				"classification": false,
				"moderation": false,
				"audio": false
			},
			"name": "mistral-ocr-2512",
			"description": "Official mistral-ocr-2512 Mistral AI model",
			"max_context_length": 16384,
			"aliases": [
				"mistral-ocr-latest"
			],
			"deprecation": null,
			"deprecation_replacement_model": null,
			"default_model_temperature": 0.0,
			"type": "base"
		},
		{
			"id": "mistral-moderation-latest",
			"object": "model",
			"created": 1766520693,
			"owned_by": "mistralai",
			"capabilities": {
				"completion_chat": false,
				"function_calling": false,
				"completion_fim": false,
				"fine_tuning": false,
				"vision": false,
				"ocr": false,
				"classification": true,
				"moderation": true,
				"audio": false
			},
			"name": "mistral-moderation-2411",
			"description": "Official mistral-moderation-2411 Mistral AI model",
			"max_context_length": 8192,
			"aliases": [
				"mistral-moderation-2411"
			],
			"deprecation": null,
			"deprecation_replacement_model": null,
			"default_model_temperature": null,
			"type": "base"
		}
	]
}
`
)

func TestClientImpl_SearchModelsByCapabilities(t *testing.T) {
	t.Run("should return models with chat completion and tool calling capabilities", func(t *testing.T) {
		// Given
		var gotReq string
		mockServer := makeMockServerWithCapture(t, "GET", "/v1/models", listModelJsonResp, http.StatusOK, &gotReq)
		defer mockServer.Close()

		ctx := context.TODO()
		c := mistral.New("fakeApiKey", mistral.WithBaseApiUrl(mockServer.URL))

		// When
		res, err := c.SearchModels(ctx, &mistral.ModelCapabilities{
			CompletionChat:  true,
			FunctionCalling: true,
		})

		// Then
		assert.NoError(t, err)
		assert.Equal(t, 3, len(res))
		assert.Equal(t, "mistral-medium-latest", res[0].Id)
		assert.Equal(t, "open-mistral-nemo", res[1].Id)
		assert.Equal(t, "devstral-small-2507", res[2].Id)
	})

	t.Run("should return models with moderation capability only", func(t *testing.T) {
		// Given
		var gotReq string
		mockServer := makeMockServerWithCapture(t, "GET", "/v1/models", listModelJsonResp, http.StatusOK, &gotReq)
		defer mockServer.Close()

		ctx := context.TODO()
		c := mistral.New("fakeApiKey", mistral.WithBaseApiUrl(mockServer.URL))

		// When
		res, err := c.SearchModels(ctx, &mistral.ModelCapabilities{
			Moderation: true,
		})

		// Then
		assert.NoError(t, err)
		assert.Equal(t, 1, len(res))
		assert.Equal(t, "mistral-moderation-latest", res[0].Id)
	})
}

func TestBaseModelCard_HasNoCapabilities(t *testing.T) {
	t.Run("should return true when model has no capabilities", func(t *testing.T) {
		// Given
		model := mistral.BaseModelCard{
			Name:         "mistral-embed",
			Id:           "mistral-embed-2507",
			Capabilities: mistral.ModelCapabilities{},
		}

		// When
		res := model.HasNoCapabilities()

		// Then
		assert.True(t, res)
	})

	t.Run("should return false when model has at least one capability", func(t *testing.T) {
		// Given
		model := mistral.BaseModelCard{
			Name: "mistral-embed",
			Id:   "mistral-embed-2507",
			Capabilities: mistral.ModelCapabilities{
				CompletionChat: true,
			},
		}

		// When
		res := model.HasNoCapabilities()

		// Then
		assert.False(t, res)
	})
}

func TestBaseModelCard_IsEmbedding(t *testing.T) {
	t.Run("should return true when model is an embedding model", func(t *testing.T) {
		// Given
		model := mistral.BaseModelCard{
			Name: "mistral-embed",
			Id:   "mistral-embed-2507",
		}

		// When
		res := model.IsEmbedding()

		// Then
		assert.True(t, res)
	})

	t.Run("should return false when model is not an embedding model", func(t *testing.T) {
		// Given
		model := mistral.BaseModelCard{
			Name: "mistral-large",
			Id:   "mistral-large-2507",
		}

		// When
		res := model.IsEmbedding()

		// Then
		assert.False(t, res)
	})
}
