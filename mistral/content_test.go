package mistral_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/mistral-client/mistral"
)

func TestTextChunk(t *testing.T) {
	t.Run("should be unmarshaled from json", func(t *testing.T) {
		j := `{"text": "hello", "type": "text"}`
		var tc mistral.TextChunk

		assert.NoError(t, json.Unmarshal([]byte(j), &tc))
		assert.Equal(t, "hello", tc.Text)
		assert.Equal(t, mistral.ContentTypeText, tc.Type())
	})

	t.Run("should be marshaled to json", func(t *testing.T) {
		tc := mistral.NewTextChunk("hello")
		j, err := json.Marshal(tc)

		assert.NoError(t, err)
		assert.Equal(t, `{"type":"text","text":"hello"}`, string(j))
	})
}

func TestImageUrlChunk(t *testing.T) {
	t.Run("should be unmarshaled from json", func(t *testing.T) {
		j := `{"image_url": "https://example.com/image.png", "type": "image_url"}`
		var ic mistral.ImageUrlChunk

		assert.NoError(t, json.Unmarshal([]byte(j), &ic))
		assert.Equal(t, "https://example.com/image.png", ic.ImageURL)
		assert.Equal(t, mistral.ContentTypeImageURL, ic.Type())
	})

	t.Run("should be marshaled to json", func(t *testing.T) {
		ic := mistral.NewImageUrlChunk("https://example.com/image.png")
		j, err := json.Marshal(ic)

		assert.NoError(t, err)
		assert.Equal(t, `{"type":"image_url","image_url":"https://example.com/image.png"}`, string(j))
	})
}

func TestDocumentUrlChunk(t *testing.T) {
	t.Run("should be unmarshaled from json", func(t *testing.T) {
		j := `{"document_name": "document.pdf", "document_url": "https://example.com/document.pdf", "type": "document_url"}`
		var duc mistral.DocumentUrlChunk

		assert.NoError(t, json.Unmarshal([]byte(j), &duc))
		assert.Equal(t, "document.pdf", duc.DocumentName)
		assert.Equal(t, "https://example.com/document.pdf", duc.DocumentURL)
		assert.Equal(t, mistral.ContentTypeDocumentURL, duc.Type())
	})

	t.Run("should be marshaled to json", func(t *testing.T) {
		duc := mistral.NewDocumentUrlChunk("document.pdf", "https://example.com/document.pdf")
		j, err := json.Marshal(duc)

		assert.NoError(t, err)
		assert.Equal(t, `{"type":"document_url","document_name":"document.pdf","document_url":"https://example.com/document.pdf"}`, string(j))
	})

	t.Run("should be unmarshaled from json with null document name", func(t *testing.T) {
		j := `{"document_name": null, "document_url": "https://example.com/document.pdf", "type": "document_url"}`
		var duc mistral.DocumentUrlChunk

		assert.NoError(t, json.Unmarshal([]byte(j), &duc))
		assert.Equal(t, "", duc.DocumentName)
		assert.Equal(t, "https://example.com/document.pdf", duc.DocumentURL)
		assert.Equal(t, mistral.ContentTypeDocumentURL, duc.Type())
	})

	t.Run("should be marshaled to json with null document name", func(t *testing.T) {
		duc := mistral.NewDocumentUrlChunk("", "https://example.com/document.pdf")
		j, err := json.Marshal(duc)

		assert.NoError(t, err)
		assert.Equal(t, `{"type":"document_url","document_url":"https://example.com/document.pdf"}`, string(j))
	})
}

func TestReferenceChunk(t *testing.T) {
	t.Run("should be unmarshaled from json", func(t *testing.T) {
		j := `{"reference_ids": [1, 2, 3, 5, 8], "type": "reference"}`
		var rc mistral.ReferenceChunk

		assert.NoError(t, json.Unmarshal([]byte(j), &rc))
		assert.Equal(t, []int{1, 2, 3, 5, 8}, rc.ReferenceIds)
		assert.Equal(t, mistral.ContentTypeReference, rc.Type())
	})

	t.Run("should be marshaled to json", func(t *testing.T) {
		rc := mistral.NewReferenceChunk(1, 2, 3, 5, 8)
		j, err := json.Marshal(rc)

		assert.NoError(t, err)
		assert.Equal(t, `{"type":"reference","reference_ids":[1,2,3,5,8]}`, string(j))
	})
}

func TestFileChunk(t *testing.T) {
	t.Run("should be unmarshaled from json", func(t *testing.T) {
		j := `{"file_id": "1234567890", "type": "file"}`
		var fc mistral.FileChunk

		assert.NoError(t, json.Unmarshal([]byte(j), &fc))
		assert.Equal(t, "1234567890", fc.FileId)
		assert.Equal(t, mistral.ContentTypeFile, fc.Type())
	})

	t.Run("should be marshaled to json", func(t *testing.T) {
		fc := mistral.NewFileChunk("1234567890")
		j, err := json.Marshal(fc)

		assert.NoError(t, err)
		assert.Equal(t, `{"type":"file","file_id":"1234567890"}`, string(j))
	})
}

func TestThinkChunk(t *testing.T) {
	t.Run("should be unmarshaled from json", func(t *testing.T) {
		j := `{
			"type": "thinking",
			"closed": false,
			"thinking": [
				{"type": "text", "text": "hello"},
				{"type": "reference", "reference_ids": [1, 2, 3]}
			]
		}`
		var tc mistral.ThinkChunk

		assert.NoError(t, json.Unmarshal([]byte(j), &tc))
		assert.Equal(t, mistral.ContentTypeThink, tc.Type())
		assert.Equal(t, false, tc.Closed)
		assert.Len(t, tc.Thinking, 2)
		assert.Equal(t, mistral.ContentTypeText, tc.Thinking[0].Type())
		assert.Equal(t, "hello", tc.Thinking[0].(*mistral.TextChunk).Text)
		assert.Equal(t, mistral.ContentTypeReference, tc.Thinking[1].Type())
		assert.Equal(t, []int{1, 2, 3}, tc.Thinking[1].(*mistral.ReferenceChunk).ReferenceIds)
	})

	t.Run("should be unmarshaled from json with closed true by default", func(t *testing.T) {
		j := `{
			"type": "thinking",
			"thinking": []
		}`
		var tc mistral.ThinkChunk

		assert.NoError(t, json.Unmarshal([]byte(j), &tc))
		assert.Equal(t, mistral.ContentTypeThink, tc.Type())
		assert.Equal(t, true, tc.Closed)
		assert.Len(t, tc.Thinking, 0)
	})

	t.Run("should be marshaled to json", func(t *testing.T) {
		tc := mistral.NewThinkChunk(
			mistral.NewTextChunk("hello"),
			mistral.NewReferenceChunk(1, 2, 3),
		)
		j, err := json.Marshal(tc)

		assert.NoError(t, err)
		assert.Equal(t, `{"type":"thinking","closed":true,"thinking":[{"type":"text","text":"hello"},{"type":"reference","reference_ids":[1,2,3]}]}`, string(j))
	})

	t.Run("should be marshaled to json with closed false", func(t *testing.T) {
		tc := mistral.NewThinkChunk()
		tc.Closed = false
		j, err := json.Marshal(tc)

		assert.NoError(t, err)
		assert.Equal(t, `{"type":"thinking","closed":false,"thinking":[]}`, string(j))
	})

	t.Run("should panic when trying to add an unsupported content type", func(t *testing.T) {
		assert.PanicsWithValue(t, "only text and reference content can be added to a thinking content", func() {
			mistral.NewThinkChunk(mistral.NewImageUrlChunk("https://example.com/image.png"))
		})
	})

	t.Run("should panic when trying to add a nil content", func(t *testing.T) {
		assert.PanicsWithValue(t, "nil content cannot be added to a thinking content", func() {
			mistral.NewThinkChunk(nil)
		})
	})
}

func TestAudioChunk(t *testing.T) {
	t.Run("should be unmarshaled from json", func(t *testing.T) {
		j := `{"input_audio": "https://example.com/audio.mp3", "type": "input_audio"}`
		var ac mistral.AudioChunk

		assert.NoError(t, json.Unmarshal([]byte(j), &ac))
		assert.Equal(t, mistral.ContentTypeAudio, ac.Type())
		assert.Equal(t, "https://example.com/audio.mp3", ac.InputAudio)
	})

	t.Run("should be marshaled to json", func(t *testing.T) {
		ac := mistral.NewAudioChunk("https://example.com/audio.mp3")
		j, err := json.Marshal(ac)

		assert.NoError(t, err)
		assert.Equal(t, `{"type":"input_audio","input_audio":"https://example.com/audio.mp3"}`, string(j))
	})
}
