package mistral

import (
	"encoding/json"
)

type ContentType string

func (c ContentType) String() string {
	return string(c)
}

const (
	ContentTypeText        ContentType = "text"
	ContentTypeImageURL    ContentType = "image_url"
	ContentTypeDocumentURL ContentType = "document_url"
	ContentTypeReference   ContentType = "reference"
	ContentTypeFile        ContentType = "file"
	ContentTypeThink       ContentType = "thinking"
	ContentTypeAudio       ContentType = "input_audio"
)

type Content interface {
	String() string
	Chunks() []ContentChunk
}

type ContentString string

var _ Content = (*ContentString)(nil)

func (s ContentString) String() string {
	return string(s)
}

func (s ContentString) Chunks() []ContentChunk {
	return nil
}

type ContentChunks []ContentChunk

var _ Content = ContentChunks{}

func (c ContentChunks) String() string {
	return ""
}

func (c ContentChunks) Chunks() []ContentChunk {
	return c
}

type ContentChunk interface {
	Type() ContentType
}

type TextChunk struct {
	ContentType ContentType `json:"type"`
	Text        string      `json:"text"`
}

var _ ContentChunk = (*TextChunk)(nil)

func NewTextChunk(text string) *TextChunk {
	return &TextChunk{
		ContentType: ContentTypeText,
		Text:        text,
	}
}

func (c *TextChunk) Type() ContentType {
	return ContentTypeText
}

type ImageUrlChunk struct {
	ContentType ContentType `json:"type"`
	ImageURL    string      `json:"image_url"`
}

var _ ContentChunk = (*ImageUrlChunk)(nil)

func NewImageUrlChunk(imageUrl string) *ImageUrlChunk {
	return &ImageUrlChunk{
		ContentType: ContentTypeImageURL,
		ImageURL:    imageUrl,
	}
}

func (c *ImageUrlChunk) Type() ContentType {
	return ContentTypeImageURL
}

type DocumentUrlChunk struct {
	ContentType  ContentType `json:"type"`
	DocumentName string      `json:"document_name,omitempty"`
	DocumentURL  string      `json:"document_url"`
}

var _ ContentChunk = (*DocumentUrlChunk)(nil)

func NewDocumentUrlChunk(documentName, documentURL string) *DocumentUrlChunk {
	return &DocumentUrlChunk{
		ContentType:  ContentTypeDocumentURL,
		DocumentName: documentName,
		DocumentURL:  documentURL,
	}
}

func (c *DocumentUrlChunk) Type() ContentType {
	return ContentTypeDocumentURL
}

type ReferenceChunk struct {
	ContentType  ContentType `json:"type"`
	ReferenceIds []int       `json:"reference_ids"`
}

var _ ContentChunk = (*ReferenceChunk)(nil)

func NewReferenceChunk(referenceIds ...int) *ReferenceChunk {
	return &ReferenceChunk{
		ContentType:  ContentTypeReference,
		ReferenceIds: referenceIds,
	}
}

func (c *ReferenceChunk) Type() ContentType {
	return ContentTypeReference
}

type FileChunk struct {
	ContentType ContentType `json:"type"`
	FileId      string      `json:"file_id"`
}

var _ ContentChunk = (*FileChunk)(nil)

func NewFileChunk(fileId string) *FileChunk {
	return &FileChunk{
		ContentType: ContentTypeFile,
		FileId:      fileId,
	}
}

func (c *FileChunk) Type() ContentType {
	return ContentTypeFile
}

type ThinkChunk struct {
	ContentType ContentType    `json:"type"`
	Closed      bool           `json:"closed"`
	Thinking    []ContentChunk `json:"thinking"`
}

var _ ContentChunk = (*ThinkChunk)(nil)
var _ json.Unmarshaler = (*ThinkChunk)(nil)

func NewThinkChunk(thinking ...ContentChunk) *ThinkChunk {
	c := &ThinkChunk{
		ContentType: ContentTypeThink,
		Closed:      true,
		Thinking:    make([]ContentChunk, 0),
	}
	for _, t := range thinking {
		if t == nil {
			panic("nil content cannot be added to a thinking content")
		}
		if t.Type() != ContentTypeText && t.Type() != ContentTypeReference {
			panic("only text and reference content can be added to a thinking content")
		}
		c.Thinking = append(c.Thinking, t)
	}
	return c
}

func (c *ThinkChunk) Type() ContentType {
	return ContentTypeThink
}

func (c *ThinkChunk) UnmarshalJSON(data []byte) error {
	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	c.ContentType = ContentType(res["type"].(string))
	thinkings := res["thinking"].([]any)
	c.Thinking = make([]ContentChunk, len(thinkings))

	for i, thinking := range thinkings {
		t := thinking.(map[string]any)

		switch t["type"].(string) {
		case ContentTypeText.String():
			c.Thinking[i] = NewTextChunk(t["text"].(string))

		case ContentTypeReference.String():
			refIds, exists := t["reference_ids"]
			var refIdsSlice []int
			if exists {
				for _, refId := range refIds.([]any) {
					refIdsSlice = append(refIdsSlice, int(refId.(float64)))
				}
			}
			c.Thinking[i] = NewReferenceChunk(refIdsSlice...)
		}
	}

	if cl, ok := res["closed"]; ok {
		c.Closed = cl.(bool)
	} else {
		c.Closed = true
	}

	return nil
}

type AudioChunk struct {
	ContentType ContentType `json:"type"`
	InputAudio  string      `json:"input_audio"`
}

var _ ContentChunk = (*AudioChunk)(nil)

// NewAudioChunk creates a new audio content chunk.
// Input audio can be:
//   - a URL to an audio file hosted on the internet
//   - a base64 encoded audio file
//   - URl of an uploaded audio file on La Plateforme
func NewAudioChunk(inputAudio string) *AudioChunk {
	return &AudioChunk{
		ContentType: ContentTypeAudio,
		InputAudio:  inputAudio,
	}
}

func (c *AudioChunk) Type() ContentType {
	return ContentTypeAudio
}
