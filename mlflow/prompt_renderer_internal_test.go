package mlflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGoTemplateRenderer_Render(t *testing.T) {
	t.Run("should render complex template with sprig functions", func(t *testing.T) {
		// Given
		renderer := &goTemplateRenderer{}
		input := `
Hello {{ .user.name | upper }}!
Welcome to {{ .project.name | default "Unknown Project" }}.
Your roles are:
{{- range .user.roles }}
- {{ . | title }}
{{- end }}
Current Date: {{ .now | date "2006-01-02" }}
Random string: {{ "hello" | shuffle }}
`
		params := map[string]any{
			"user": map[string]any{
				"name":  "john doe",
				"roles": []string{"admin", "editor"},
			},
			"project": map[string]any{
				"name": "Mistral Client",
			},
			"now": "2026-02-02T10:00:00Z",
		}

		// When
		res, err := renderer.Render(input, params)

		// Then
		require.NoError(t, err)
		assert.Contains(t, res, `Hello JOHN DOE!
Welcome to Mistral Client.
Your roles are:
- Admin
- Editor
Current Date: 2026-02-02`)
		assert.NotEqual(t, "hello", res) // shuffle should have changed it
	})

	t.Run("should return error for invalid template syntax", func(t *testing.T) {
		// Given
		renderer := &goTemplateRenderer{}
		input := "Hello {{ .name" // Missing closing braces

		// When
		res, err := renderer.Render(input, nil)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse template")
		assert.Empty(t, res)
	})

	t.Run("should return error when parameter is missing and missingkey=error is set", func(t *testing.T) {
		// Given
		renderer := &goTemplateRenderer{}
		input := "Hello {{ .name }}!"
		params := map[string]any{} // name is missing

		// When
		res, err := renderer.Render(input, params)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "map has no entry for key \"name\"")
		assert.Empty(t, res)
	})

	t.Run("should return error when function execution fails", func(t *testing.T) {
		// Given
		renderer := &goTemplateRenderer{}
		input := "{{ fail \"custom error\" }}"
		params := map[string]any{}

		// When
		res, err := renderer.Render(input, params)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "custom error")
		assert.Empty(t, res)
	})

	t.Run("should handle nil params", func(t *testing.T) {
		// Given
		renderer := &goTemplateRenderer{}
		input := "Static content"

		// When
		res, err := renderer.Render(input, nil)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, "Static content", res)
	})
}

func TestRawRenderer_Render(t *testing.T) {
	t.Run("should return raw content as is", func(t *testing.T) {
		// Given
		renderer := &rawRenderer{}
		input := "Static content {{ name }}, {{ .User.Name }}"

		// When
		res, err := renderer.Render(input, nil)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, input, res)
	})
}
