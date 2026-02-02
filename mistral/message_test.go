package mistral_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/mistral-client/mistral"
	"github.com/thomas-marquis/mistral-client/mlflow"
	"github.com/thomas-marquis/mistral-client/mocks"
	"go.uber.org/mock/gomock"
)

func TestSystemMessage(t *testing.T) {
	t.Run("should be unmarshaled from json with string content", func(t *testing.T) {
		j := `{"role": "system", "content": "hello"}`
		var sm mistral.SystemMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &sm))
		assert.Equal(t, mistral.ContentString("hello"), sm.MessageContent)
		assert.Equal(t, mistral.RoleSystem, sm.Role())
		assert.Equal(t, mistral.RoleSystem, sm.MessageRole)
	})

	t.Run("should be unmarshaled from json with array content", func(t *testing.T) {
		j := `{
			"role": "system", 
			"content": [
				{"type": "text", "text": "hello"},
				{"type": "thinking", "thinking": [{"type": "text", "text": "world"}]}
			]
		}`
		var sm mistral.SystemMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &sm))
		assert.Equal(t, mistral.RoleSystem, sm.MessageRole)
		assert.Len(t, sm.MessageContent.Chunks(), 2)
		assert.Equal(t, "hello", sm.MessageContent.Chunks()[0].(*mistral.TextChunk).Text)
		assert.Equal(t, "world", sm.MessageContent.Chunks()[1].(*mistral.ThinkChunk).Thinking[0].(*mistral.TextChunk).Text)
	})

	t.Run("should be marshaled to json with simple string content", func(t *testing.T) {
		sm := mistral.NewSystemMessage(mistral.ContentString("hello"))
		j, err := json.Marshal(sm)

		assert.NoError(t, err)
		assert.Equal(t, `{"role":"system","content":"hello"}`, string(j))
	})

	t.Run("should be marshaled to json with chunked content", func(t *testing.T) {
		sm := mistral.NewSystemMessage(
			mistral.ContentChunks{
				mistral.NewTextChunk("hello"),
				mistral.NewThinkChunk(mistral.NewTextChunk("world")),
			},
		)
		j, err := json.Marshal(sm)

		assert.NoError(t, err)
		assert.Equal(t, `{"role":"system","content":[{"type":"text","text":"hello"},{"type":"thinking","closed":true,"thinking":[{"type":"text","text":"world"}]}]}`, string(j))
	})
}

func TestUserMessage(t *testing.T) {
	t.Run("should be unmarshaled from json with chunked content", func(t *testing.T) {
		j := `{
			"role": "user", 
			"content": [
				{"type": "text", "text": "hello"},
				{"image_url": "https://example.com/image.png", "type": "image_url"},
				{"document_name": "document.pdf", "document_url": "https://example.com/document.pdf", "type": "document_url"},
				{"reference_ids": [1, 2, 3, 5, 8], "type": "reference"},
				{"file_id": "1234567890", "type": "file"},
				{"type": "thinking", "thinking": [{"type": "text", "text": "world"}]},
				{"input_audio": "https://example.com/audio.mp3", "type": "input_audio"}
			]
		}`
		var um mistral.UserMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &um))
		assert.Equal(t, mistral.RoleUser, um.MessageRole)
		assert.Len(t, um.MessageContent.Chunks(), 7)

		assert.Equal(t, "hello", um.MessageContent.Chunks()[0].(*mistral.TextChunk).Text)
		assert.Equal(t, "https://example.com/image.png", um.MessageContent.Chunks()[1].(*mistral.ImageUrlChunk).ImageURL)
		assert.Equal(t, "document.pdf", um.MessageContent.Chunks()[2].(*mistral.DocumentUrlChunk).DocumentName)
		assert.Equal(t, "https://example.com/document.pdf", um.MessageContent.Chunks()[2].(*mistral.DocumentUrlChunk).DocumentURL)
		assert.Equal(t, []int{1, 2, 3, 5, 8}, um.MessageContent.Chunks()[3].(*mistral.ReferenceChunk).ReferenceIds)
		assert.Equal(t, "1234567890", um.MessageContent.Chunks()[4].(*mistral.FileChunk).FileId)
		assert.Equal(t, "world", um.MessageContent.Chunks()[5].(*mistral.ThinkChunk).Thinking[0].(*mistral.TextChunk).Text)
		assert.Equal(t, "https://example.com/audio.mp3", um.MessageContent.Chunks()[6].(*mistral.AudioChunk).InputAudio)

	})

	t.Run("should be unmarshalled from json with simple string content", func(t *testing.T) {
		j := `{"role": "user", "content": "hello"}`
		var um mistral.UserMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &um))
		assert.Equal(t, mistral.ContentString("hello"), um.MessageContent)
		assert.Equal(t, mistral.RoleUser, um.MessageRole)
	})

	t.Run("should be unmarshalled from json with null content", func(t *testing.T) {
		j := `{"role": "user", "content": null}`
		var um mistral.UserMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &um))
		assert.Equal(t, nil, um.MessageContent)
		assert.Equal(t, mistral.RoleUser, um.MessageRole)
	})

	t.Run("should be unmarshalled from json with omitted content", func(t *testing.T) {
		j := `{"role": "user"}`
		var um mistral.UserMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &um))
		assert.Equal(t, nil, um.MessageContent)
		assert.Equal(t, mistral.RoleUser, um.MessageRole)
	})

	t.Run("should be marshaled to json with simple string content", func(t *testing.T) {
		um := mistral.NewUserMessage(mistral.ContentString("hello"))
		j, err := json.Marshal(um)

		assert.NoError(t, err)
		assert.Equal(t, `{"role":"user","content":"hello"}`, string(j))
	})

	t.Run("should be marshaled to json with chunked content", func(t *testing.T) {
		um := mistral.NewUserMessage(
			mistral.ContentChunks{
				mistral.NewTextChunk("hello"),
				mistral.NewThinkChunk(mistral.NewTextChunk("world")),
			},
		)
		j, err := json.Marshal(um)

		assert.NoError(t, err)
		assert.Equal(t, `{"role":"user","content":[{"type":"text","text":"hello"},{"type":"thinking","closed":true,"thinking":[{"type":"text","text":"world"}]}]}`, string(j))
	})
}

func TestAssistantMessage(t *testing.T) {
	t.Run("should be unmarshaled from json with chunked content", func(t *testing.T) {
		j := `{
			"role": "assistant", 
			"tool_calls": null,
			"content": [
				{"type": "text", "text": "hello"},
				{"image_url": "https://example.com/image.png", "type": "image_url"},
				{"document_name": "document.pdf", "document_url": "https://example.com/document.pdf", "type": "document_url"},
				{"reference_ids": [1, 2, 3, 5, 8], "type": "reference"},
				{"file_id": "1234567890", "type": "file"},
				{"type": "thinking", "thinking": [{"type": "text", "text": "world"}]},
				{"input_audio": "https://example.com/audio.mp3", "type": "input_audio"}
			]
		}`
		var am mistral.AssistantMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &am))
		assert.Equal(t, mistral.RoleAssistant, am.MessageRole)
		assert.Equal(t, mistral.RoleAssistant, am.Role())
		assert.Equal(t, false, am.Prefix)
		assert.Nil(t, am.ToolCalls)
		assert.Len(t, am.MessageContent.Chunks(), 7)

		assert.Equal(t, "hello", am.MessageContent.Chunks()[0].(*mistral.TextChunk).Text)
		assert.Equal(t, "https://example.com/image.png", am.MessageContent.Chunks()[1].(*mistral.ImageUrlChunk).ImageURL)
		assert.Equal(t, "document.pdf", am.MessageContent.Chunks()[2].(*mistral.DocumentUrlChunk).DocumentName)
		assert.Equal(t, "https://example.com/document.pdf", am.MessageContent.Chunks()[2].(*mistral.DocumentUrlChunk).DocumentURL)
		assert.Equal(t, []int{1, 2, 3, 5, 8}, am.MessageContent.Chunks()[3].(*mistral.ReferenceChunk).ReferenceIds)
		assert.Equal(t, "1234567890", am.MessageContent.Chunks()[4].(*mistral.FileChunk).FileId)
		assert.Equal(t, "world", am.MessageContent.Chunks()[5].(*mistral.ThinkChunk).Thinking[0].(*mistral.TextChunk).Text)
		assert.Equal(t, "https://example.com/audio.mp3", am.MessageContent.Chunks()[6].(*mistral.AudioChunk).InputAudio)
	})

	t.Run("should be unmarshable with tool calls", func(t *testing.T) {
		j := `{
			"role": "assistant", 
			"content": [
				{"type": "text", "text": "hello"}
			],
			"tool_calls": [
				{
					"type": "function",
					"id": "123",
					"index": 0,
					"function": {
						"arguments": {"name": "toto"},
						"name": "testFunction"
					}
				}	
			]
		}`
		var am mistral.AssistantMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &am))
		assert.Equal(t, mistral.RoleAssistant, am.MessageRole)
		assert.Len(t, am.ToolCalls, 1)
		assert.Equal(t, "123", am.ToolCalls[0].ID)
		assert.Equal(t, "testFunction", am.ToolCalls[0].Function.Name)
	})

	t.Run("should be unmarshalled from json with simple string content", func(t *testing.T) {
		j := `{"role": "assistant", "content": "hello"}`
		var am mistral.AssistantMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &am))
		assert.Equal(t, mistral.ContentString("hello"), am.MessageContent)
		assert.Equal(t, mistral.RoleAssistant, am.MessageRole)
	})

	t.Run("should be unmarshalled from json with null content", func(t *testing.T) {
		j := `{"role": "assistant", "content": null}`
		var am mistral.AssistantMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &am))
		assert.Equal(t, nil, am.MessageContent)
		assert.Equal(t, mistral.RoleAssistant, am.MessageRole)
	})

	t.Run("should be unmarshalled from json with omitted content", func(t *testing.T) {
		j := `{"role": "assistant"}`
		var am mistral.AssistantMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &am))
		assert.Equal(t, nil, am.MessageContent)
		assert.Equal(t, mistral.RoleAssistant, am.MessageRole)
	})

	t.Run("should be marshaled to json with simple string content", func(t *testing.T) {
		am := mistral.NewAssistantMessage(mistral.ContentString("hello"))
		j, err := json.Marshal(am)

		assert.NoError(t, err)
		assert.Equal(t, `{"role":"assistant","content":"hello"}`, string(j))
	})

	t.Run("should be marshaled to json with chunked content", func(t *testing.T) {
		am := mistral.NewAssistantMessage(
			mistral.ContentChunks{
				mistral.NewTextChunk("hello"),
				mistral.NewThinkChunk(mistral.NewTextChunk("world")),
			},
		)
		j, err := json.Marshal(am)

		assert.NoError(t, err)
		assert.Equal(t, `{"role":"assistant","content":[{"type":"text","text":"hello"},{"type":"thinking","closed":true,"thinking":[{"type":"text","text":"world"}]}]}`, string(j))
	})

	t.Run("should be marshaled to json with tool calls", func(t *testing.T) {
		am := mistral.NewAssistantMessage(
			mistral.ContentString("coucou"),
			mistral.NewToolCall("123", 0, "testFunction", map[string]interface{}{"name": "toto"}),
		)
		j, err := json.Marshal(am)

		assert.NoError(t, err)
		assert.Equal(t, `{"role":"assistant","content":"coucou","tool_calls":[{"id":"123","index":0,"function":{"name":"testFunction","arguments":{"input":{"name":"toto"}}},"type":"function"}]}`, string(j))
	})
}

func TestToolMessage(t *testing.T) {
	t.Run("should be unmarshaled from json", func(t *testing.T) {
		j := `{
			"role": "tool",
			"content": [
				{"type": "text", "text": "hello"},
				{"image_url": "https://example.com/image.png", "type": "image_url"},
				{"document_name": "document.pdf", "document_url": "https://example.com/document.pdf", "type": "document_url"},
				{"reference_ids": [1, 2, 3, 5, 8], "type": "reference"},
				{"file_id": "1234567890", "type": "file"},
				{"type": "thinking", "thinking": [{"type": "text", "text": "world"}]},
				{"input_audio": "https://example.com/audio.mp3", "type": "input_audio"}
			],
			"name": "testFunction",
			"tool_call_id": "azerty"
		}`
		var tm mistral.ToolMessage

		assert.NoError(t, json.Unmarshal([]byte(j), &tm))
		assert.Equal(t, mistral.RoleTool, tm.MessageRole)
		assert.Equal(t, mistral.RoleTool, tm.Role())

		assert.Equal(t, "testFunction", tm.Name)
		assert.Equal(t, "azerty", tm.ToolCallId)

		assert.Len(t, tm.MessageContent.Chunks(), 7)
		assert.Equal(t, "hello", tm.MessageContent.Chunks()[0].(*mistral.TextChunk).Text)
		assert.Equal(t, "https://example.com/image.png", tm.MessageContent.Chunks()[1].(*mistral.ImageUrlChunk).ImageURL)
		assert.Equal(t, "document.pdf", tm.MessageContent.Chunks()[2].(*mistral.DocumentUrlChunk).DocumentName)
		assert.Equal(t, "https://example.com/document.pdf", tm.MessageContent.Chunks()[2].(*mistral.DocumentUrlChunk).DocumentURL)
		assert.Equal(t, []int{1, 2, 3, 5, 8}, tm.MessageContent.Chunks()[3].(*mistral.ReferenceChunk).ReferenceIds)
		assert.Equal(t, "1234567890", tm.MessageContent.Chunks()[4].(*mistral.FileChunk).FileId)
		assert.Equal(t, "world", tm.MessageContent.Chunks()[5].(*mistral.ThinkChunk).Thinking[0].(*mistral.TextChunk).Text)
		assert.Equal(t, "https://example.com/audio.mp3", tm.MessageContent.Chunks()[6].(*mistral.AudioChunk).InputAudio)
	})

	t.Run("should be marshaled to json", func(t *testing.T) {
		tm := mistral.NewToolMessage(
			"testFunction",
			"azerty",
			mistral.ContentChunks{
				mistral.NewTextChunk("hello"),
				mistral.NewThinkChunk(mistral.NewTextChunk("world")),
			},
		)
		j, err := json.Marshal(tm)

		assert.NoError(t, err)
		assert.Equal(t, `{"role":"tool","content":[{"type":"text","text":"hello"},{"type":"thinking","closed":true,"thinking":[{"type":"text","text":"world"}]}],"name":"testFunction","tool_call_id":"azerty"}`, string(j))
	})
}

func TestMessagesFromRegisteredPrompt(t *testing.T) {
	t.Run("should return a rendered single user message when the prompt is a text", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockRegistry := mocks.NewMockPromptRegistry(ctrl)

		fakePrompt := mlflow.NewPromptText("my_prompt", mlflow.VersionLatest, "Say {{message}}")

		mockRegistry.EXPECT().
			Get(gomock.AssignableToTypeOf(ctxType), gomock.Eq("my_prompt"), gomock.Eq(mlflow.VersionLatest)).
			Return(fakePrompt, nil).
			Times(1)

		// When
		res, err := mistral.MessagesFromRegisteredPrompt(context.TODO(),
			mockRegistry, "my_prompt", mlflow.VersionLatest,
			map[string]any{"message": "hello"})

		// Then
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, mistral.RoleUser, res[0].Role())
		assert.Equal(t, "Say hello", res[0].Content().String())
	})

	t.Run("should return a rendered all messages when the prompt is a chat", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockRegistry := mocks.NewMockPromptRegistry(ctrl)

		fakePrompt := mlflow.NewPromptChat("my_prompt", mlflow.VersionLatest,
			[]mlflow.PromptChatMessage{
				{Role: mlflow.PromptRoleUser, Content: "You are an experienced {{occupation}}"},
				{Role: mlflow.PromptRoleAssistant, Content: "Please answer this question: {{question}}"},
			})

		mockRegistry.EXPECT().
			Get(gomock.AssignableToTypeOf(ctxType), gomock.Eq("my_prompt"), gomock.Eq(mlflow.VersionLatest)).
			Return(fakePrompt, nil).
			Times(1)

		// When
		res, err := mistral.MessagesFromRegisteredPrompt(context.TODO(),
			mockRegistry, "my_prompt", mlflow.VersionLatest,
			map[string]any{"occupation": "writer", "question": "What is your favourite book?"})

		// Then
		assert.NoError(t, err)
		assert.Len(t, res, 2)

		assert.Equal(t, mistral.RoleUser, res[0].Role())
		assert.Equal(t, "You are an experienced writer", res[0].Content().String())

		assert.Equal(t, mistral.RoleAssistant, res[1].Role())
		assert.Equal(t, "Please answer this question: What is your favourite book?", res[1].Content().String())
	})

	t.Run("should return an error when prompt registry returns an error", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockRegistry := mocks.NewMockPromptRegistry(ctrl)

		fakeErr := errors.New("fake error")

		mockRegistry.EXPECT().
			Get(gomock.AssignableToTypeOf(ctxType), gomock.Eq("my_prompt"), gomock.Eq(mlflow.VersionLatest)).
			Return(nil, fakeErr).
			Times(1)

		// When
		_, err := mistral.MessagesFromRegisteredPrompt(context.TODO(), mockRegistry, "my_prompt", mlflow.VersionLatest, nil)

		// Then
		assert.Equal(t, fakeErr, err)
	})

	t.Run("should return an error when text prompt rendering fails", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockRegistry := mocks.NewMockPromptRegistry(ctrl)

		fakePrompt := mlflow.NewPromptText("my_prompt", mlflow.VersionLatest, "Say {{message}}")

		mockRegistry.EXPECT().
			Get(gomock.AssignableToTypeOf(ctxType), gomock.Eq("my_prompt"), gomock.Eq(mlflow.VersionLatest)).
			Return(fakePrompt, nil).
			Times(1)

		// When
		// Missing "message" parameter should trigger a rendering error
		_, err := mistral.MessagesFromRegisteredPrompt(context.TODO(),
			mockRegistry, "my_prompt", mlflow.VersionLatest,
			map[string]any{})

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing parameter: message")
	})

	t.Run("should return an error when chat prompt rendering fails", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockRegistry := mocks.NewMockPromptRegistry(ctrl)

		fakePrompt := mlflow.NewPromptChat("my_prompt", mlflow.VersionLatest,
			[]mlflow.PromptChatMessage{
				{Role: mlflow.PromptRoleUser, Content: "Hello {{name}}"},
			})

		mockRegistry.EXPECT().
			Get(gomock.AssignableToTypeOf(ctxType), gomock.Eq("my_prompt"), gomock.Eq(mlflow.VersionLatest)).
			Return(fakePrompt, nil).
			Times(1)

		// When
		// Missing "name" parameter should trigger a rendering error
		_, err := mistral.MessagesFromRegisteredPrompt(context.TODO(),
			mockRegistry, "my_prompt", mlflow.VersionLatest,
			map[string]any{})

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing parameter: name")
	})
}
