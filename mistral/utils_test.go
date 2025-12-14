package mistral_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/mistral-client/mistral"
)

func TestJsonMap(t *testing.T) {
	t.Run("should unmarshall simple json map", func(t *testing.T) {
		j := `{ "key": "value" }`
		var m mistral.JsonMap

		assert.NoError(t, json.Unmarshal([]byte(j), &m))
		assert.Equal(t, mistral.JsonMap{"key": "value"}, m)
	})

	t.Run("should unmarshall json map from quoted json string", func(t *testing.T) {
		j := `"{\"key\": \"value\"}"`
		var m mistral.JsonMap

		assert.NoError(t, json.Unmarshal([]byte(j), &m))
		assert.Equal(t, mistral.JsonMap{"key": "value"}, m)
	})

	t.Run("should unmarshall empty json map", func(t *testing.T) {
		j := `{}`
		var m mistral.JsonMap

		assert.NoError(t, json.Unmarshal([]byte(j), &m))
		assert.Equal(t, mistral.JsonMap{}, m)
	})

	t.Run("should unmarshall null json map", func(t *testing.T) {
		j := "null"
		var m mistral.JsonMap

		assert.NoError(t, json.Unmarshal([]byte(j), &m))
		assert.Equal(t, mistral.JsonMap(nil), m)
	})
}

func formatJSON(j string) string {
	var out map[string]any
	if err := json.Unmarshal([]byte(j), &out); err != nil {
		return j
	}
	return j
}
