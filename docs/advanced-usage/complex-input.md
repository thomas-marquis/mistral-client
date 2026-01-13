# Complex input data

You can send more than just a simple string to the model for chat completion.
But to do so, you need to use a chunked content. 
If you're not yet familiar with this concept, check [this page](../concepts/content-types.md).

If you want to use multimodal features (like images or audio), or if you want to provide more complex content, you can use chunks.

```go
mistral.NewUserMessage(mistral.ContentChunks{
    mistral.NewTextChunk("Describe this image:"),
    mistral.NewImageUrlChunk("https://example.com/image.jpg"), // (1)
})
```

1. The chunks order matters! Try different combinations to see what works best for you.

## Supported chunks

### Text

A simple text block. This is the most common chunk type.

```go
chunk := mistral.NewTextChunk("Hello world")
```

### Image URL

A link to an image that the model should process.

```go
chunk := mistral.NewImageUrlChunk("https://example.com/image.jpg")
```

### Audio

A base64 encoded audio string or a link to an audio file.

```go
chunk := mistral.NewAudioChunk("base64_encoded_audio_data")
```

### Document URL

A link to a document (like a PDF). You must provide both a name and a URL.

```go
chunk := mistral.NewDocumentUrlChunk("report.pdf", "https://example.com/report.pdf")
```

### File

A reference to a file already uploaded to Mistral.

```go
chunk := mistral.NewFileChunk("file_id_123456")
```

### Reference

A list of reference IDs to be used by the model.

```go
chunk := mistral.NewReferenceChunk(1, 2, 3)
```

### Thinking

Represents a "thinking" process, often used by reasoning models. It can contain other chunks (usually text).

```go
chunk := mistral.NewThinkChunk(
    mistral.NewTextChunk("I am thinking about the answer..."),
)
```