package mistral

import (
	"encoding/json"
	"errors"
	"fmt"
)

type ChatMessage interface {
	Role() Role
	Content() Content
}

type BaseMessage struct {
	MessageRole    Role    `json:"role"`
	MessageContent Content `json:"content"`
}

func (m *BaseMessage) Content() Content {
	return m.MessageContent
}

func (m *BaseMessage) Role() Role {
	return m.MessageRole
}

type SystemMessage struct {
	BaseMessage
}

var _ ChatMessage = (*SystemMessage)(nil)
var _ json.Unmarshaler = (*SystemMessage)(nil)

func NewSystemMessage(content Content) *SystemMessage {
	m := &SystemMessage{
		BaseMessage{
			MessageRole:    RoleSystem,
			MessageContent: content,
		},
	}
	return m
}

func NewSystemMessageFromString(content string) *SystemMessage {
	m := &SystemMessage{
		BaseMessage{
			MessageRole:    RoleSystem,
			MessageContent: ContentString(content),
		},
	}
	return m
}

func (m *SystemMessage) UnmarshalJSON(data []byte) error {
	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	m.MessageRole = Role(res["role"].(string))

	if stringContent, ok := res["content"].(string); ok {
		m.MessageContent = ContentString(stringContent)
	} else if listContent, ok := res["content"].([]any); ok {
		cts := ContentChunks{}
		for _, chunk := range listContent {
			t := chunk.(map[string]any)

			switch t["type"].(string) {
			case ContentTypeText.String():
				cts = append(cts, NewTextChunk(t["text"].(string)))

			case ContentTypeThink.String():
				var tc ThinkChunk
				if err := mapToStruct(t, &tc); err != nil {
					return err
				}
				cts = append(cts, &tc)
			}
		}

		m.MessageContent = cts
	} else {
		return fmt.Errorf("invalid content type: %T", res["content"])
	}

	return nil
}

type UserMessage struct {
	BaseMessage
}

var _ ChatMessage = (*UserMessage)(nil)
var _ json.Unmarshaler = (*UserMessage)(nil)

func NewUserMessage(content Content) *UserMessage {
	return &UserMessage{
		BaseMessage{
			MessageRole:    RoleUser,
			MessageContent: content,
		},
	}
}

func NewUserMessageFromString(content string) *UserMessage {
	return &UserMessage{
		BaseMessage{
			MessageRole:    RoleUser,
			MessageContent: ContentString(content),
		},
	}
}

func (m *UserMessage) UnmarshalJSON(data []byte) error {
	var res map[string]any
	var err error
	if err = json.Unmarshal(data, &res); err != nil {
		return err
	}
	m.MessageContent, err = unmarshalMessageContent(res)
	if err != nil {
		return err
	}

	m.MessageRole = Role(res["role"].(string))

	return nil
}

type AssistantMessage struct {
	BaseMessage
	Prefix    bool       `json:"prefix,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

var _ ChatMessage = (*AssistantMessage)(nil)
var _ json.Unmarshaler = (*AssistantMessage)(nil)

func NewAssistantMessage(content Content, toolCalls ...ToolCall) *AssistantMessage {
	return &AssistantMessage{
		BaseMessage: BaseMessage{
			MessageRole:    RoleAssistant,
			MessageContent: content,
		},
		ToolCalls: toolCalls,
	}
}

func NewAssistantMessageFromString(content string, toolCalls ...ToolCall) *AssistantMessage {
	return &AssistantMessage{
		BaseMessage: BaseMessage{
			MessageRole:    RoleAssistant,
			MessageContent: ContentString(content),
		},
		ToolCalls: toolCalls,
	}
}

func (m *AssistantMessage) UnmarshalJSON(data []byte) error {
	var res map[string]any
	var err error
	if err = json.Unmarshal(data, &res); err != nil {
		return err
	}
	m.MessageContent, err = unmarshalMessageContent(res)
	if err != nil {
		return err
	}

	m.MessageRole = RoleAssistant

	if tcs, ok := res["tool_calls"]; ok && tcs != nil {
		m.ToolCalls = make([]ToolCall, len(tcs.([]any)))
		for i, tc := range tcs.([]any) {
			m.ToolCalls[i] = ToolCall{}
			if err := mapToStruct(tc.(map[string]any), &m.ToolCalls[i]); err != nil {
				return err
			}
		}
	}

	return nil
}

// Output deserializes the message JSON content to the target pointer.
// Returns an error if the message content is not compatible with the target type.
// Use this method ONLY if you are sure the model answered a JSON.
// You can use one of these two options in the request:
//   - mistral.WithResponseJsonObjectFormat() + specifying the desired structure in your prompt
//   - or mistral.WithResponseJsonSchema(propertyDef)
func (m *AssistantMessage) Output(target any) error {
	c := m.MessageContent.String()
	if c == "" {
		return errors.New("unmarshalling impossible, the message content is empty")
	}
	return json.Unmarshal([]byte(c), target)
}

type ToolMessage struct {
	BaseMessage
	Name       string `json:"name"`
	ToolCallId string `json:"tool_call_id"`
}

var _ ChatMessage = (*ToolMessage)(nil)
var _ json.Unmarshaler = (*ToolMessage)(nil)

func NewToolMessage(name string, toolCallId string, content Content) *ToolMessage {
	return &ToolMessage{
		BaseMessage: BaseMessage{
			MessageRole:    RoleTool,
			MessageContent: content,
		},
		Name:       name,
		ToolCallId: toolCallId,
	}
}

func (m *ToolMessage) UnmarshalJSON(data []byte) error {
	var res map[string]any
	var err error
	if err = json.Unmarshal(data, &res); err != nil {
		return err
	}
	m.MessageContent, err = unmarshalMessageContent(res)
	if err != nil {
		return err
	}

	m.MessageRole = Role(res["role"].(string))
	m.Name = res["name"].(string)
	m.ToolCallId = res["tool_call_id"].(string)

	return nil
}

func unmarshalMessageContent(raw map[string]any) (Content, error) {
	var ct Content

	rawContent, exists := raw["content"]
	if !exists || rawContent == nil {
		return nil, nil
	}

	if stringContent, ok := rawContent.(string); ok {
		ct = ContentString(stringContent)
	} else if listContent, ok := rawContent.([]any); ok {
		cts := ContentChunks{}
		for _, chunk := range listContent {
			t := chunk.(map[string]any)
			var ptr ContentChunk

			switch t["type"].(string) {
			case ContentTypeText.String():
				ptr = &TextChunk{}
			case ContentTypeImageURL.String():
				ptr = &ImageUrlChunk{}
			case ContentTypeDocumentURL.String():
				ptr = &DocumentUrlChunk{}
			case ContentTypeReference.String():
				ptr = &ReferenceChunk{}
			case ContentTypeFile.String():
				ptr = &FileChunk{}
			case ContentTypeThink.String():
				ptr = &ThinkChunk{}
			case ContentTypeAudio.String():
				ptr = &AudioChunk{}
			}
			if err := mapToStruct(t, ptr); err != nil {
				return nil, err
			}
			cts = append(cts, ptr)
		}

		ct = cts
	} else {
		return nil, fmt.Errorf("invalid content type: %T", raw["content"])
	}
	return ct, nil
}

func mapToMessage(data map[string]any) (ChatMessage, error) {
	role, ok := data["role"].(string)
	if !ok {
		return nil, errors.New("role not found")
	}
	switch role {
	case RoleSystem.String():
		var m SystemMessage
		err := mapToStruct(data, &m)
		return &m, err
	case RoleUser.String():
		var m UserMessage
		err := mapToStruct(data, &m)
		return &m, err
	case RoleAssistant.String():
		var m AssistantMessage
		err := mapToStruct(data, &m)
		return &m, err
	case RoleTool.String():
		var m ToolMessage
		err := mapToStruct(data, &m)
		return &m, err
	default:
		return nil, errors.New("unsupported role")
	}
}
