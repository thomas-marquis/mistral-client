package mistral_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/mistral-client/mistral"
)

func TestApiError(t *testing.T) {
	t.Run("should format error message for ApiError", func(t *testing.T) {
		err := mistral.ApiError{
			Code: http.StatusBadRequest,
			Content: map[string]any{
				"object": "error",
				"message": map[string]any{
					"detail": []map[string]any{
						{
							"type":  "extra_forbidden",
							"loc":   []string{"body", "parallel_tool_calls"},
							"msg":   "Extra inputs are not permitted",
							"input": true,
						},
					},
				},
				"type":  "invalid_request_error",
				"param": nil,
				"code":  nil,
			},
		}

		assert.Equal(t, "[400] invalid_request_error: extra_forbidden: Extra inputs are not permitted (body.parallel_tool_calls)", err.Error())
	})

	t.Run("should format simple error message for ApiError (401 case)", func(t *testing.T) {
		err := mistral.ApiError{
			Code: http.StatusUnauthorized,
			Content: map[string]any{
				"detail": "Unauthorized",
			},
		}

		assert.Equal(t, "[401] Unauthorized", err.Error())
	})

	t.Run("should format error message with multiple details", func(t *testing.T) {
		err := mistral.ApiError{
			Code: http.StatusBadRequest,
			Content: map[string]any{
				"object": "error",
				"message": map[string]any{
					"detail": []map[string]any{
						{
							"type":  "extra_forbidden",
							"loc":   []string{"body", "parallel_tool_calls"},
							"msg":   "Extra inputs are not permitted",
							"input": true,
						},
						{
							"type":  "missing_required",
							"loc":   []string{"body", "messages"},
							"msg":   "Missing required property: messages",
							"input": false,
						},
					},
				},
				"type":  "invalid_request_error",
				"param": nil,
				"code":  nil,
			},
		}

		assert.Equal(t, "[400] invalid_request_error: extra_forbidden: Extra inputs are not permitted (body.parallel_tool_calls); missing_required: Missing required property: messages (body.messages)", err.Error())
	})

	t.Run("should format ApiError from json with array message", func(t *testing.T) {
		j := `{
			"object": "error",
			"message": {
				"detail": [
					{
						"type": "extra_forbidden",
						"loc": [
							"body",
							"parallel_tool_calls"
						],
						"msg": "Extra inputs are not permitted",
						"input": true
					},
					{
						"type": "missing_required",
						"loc": [
							"body",
							"messages"
						],
						"msg": "Missing required property: messages",
						"input": false
					}
				]
			},
			"type": "invalid_request_error",
			"param": null,
			"code": null
		}`

		var content map[string]any
		assert.NoError(t, json.Unmarshal([]byte(j), &content))

		err := mistral.ApiError{
			Code:    http.StatusBadRequest,
			Content: content,
		}

		assert.Equal(t, "[400] invalid_request_error: extra_forbidden: Extra inputs are not permitted (body.parallel_tool_calls); missing_required: Missing required property: messages (body.messages)", err.Error())
	})

	t.Run("should format ApiError from json with string message", func(t *testing.T) {
		j := `{
			"object": "error",
			"message": "This model does not support output_dimension.",
			"type": "invalid_request_invalid_args",
			"param": null,
			"code": "3051"
		}`
		var content map[string]any
		assert.NoError(t, json.Unmarshal([]byte(j), &content))

		err := mistral.ApiError{
			Code:    http.StatusBadRequest,
			Content: content,
		}

		assert.Equal(t, "[400] invalid_request_invalid_args: This model does not support output_dimension.", err.Error())
	})
}
