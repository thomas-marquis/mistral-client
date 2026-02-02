package mlflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleFormatPromptRenderer_Render(t *testing.T) {
	t.Run("should replace double brackets with params", func(t *testing.T) {
		// Given
		renderer := &simpleFormatRenderer{}

		input := "Hello {{user.name}}, my name is {{ me }} and I am {{ age}} years old."

		params := map[string]any{
			"user.name": "John",
			"me":        "Bob",
			"age":       30,
		}

		// When
		res, err := renderer.Render(input, params)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, "Hello John, my name is Bob and I am 30 years old.", res)
	})

	t.Run("should return an error when parameter is missing", func(t *testing.T) {
		// Given
		renderer := &simpleFormatRenderer{}
		input := "Hello {{name}}, welcome to {{city}}."
		params := map[string]any{
			"name": "John",
		}

		// When
		res, err := renderer.Render(input, params)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing parameter: city")
		assert.Empty(t, res)
	})

	t.Run("should not return error when a parameter is not used in the template string", func(t *testing.T) {
		// Given
		renderer := &simpleFormatRenderer{}
		input := "Hello {{name}}."
		params := map[string]any{
			"name":   "John",
			"unused": "value",
		}

		// When
		res, err := renderer.Render(input, params)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, "Hello John.", res)
	})

	t.Run("should return an non supported error when nested parameter maps are used", func(t *testing.T) {
		// Given
		renderer := &simpleFormatRenderer{}
		input := "Hello {{user.name}}."
		params := map[string]any{
			"user": map[string]any{
				"name": "John",
			},
		}

		// When
		res, err := renderer.Render(input, params)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nested parameters are not supported: user")
		assert.Empty(t, res)
	})
}
