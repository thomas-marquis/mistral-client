package mlflow

type PromptRole string

type PromptType string

const (
	PromptTypeText PromptType = "text"
	PromptTypeChat PromptType = "chat"

	PromptRoleSystem    PromptRole = "system"
	PromptRoleUser      PromptRole = "user"
	PromptRoleAssistant PromptRole = "assistant"
)

type Prompt interface {
	Type() PromptType
	Name() string
	Version() Version
}

var (
	_ Prompt = (*PromptText)(nil)
	_ Prompt = (*PromptChat)(nil)
)

type basePrompt struct {
	name    string
	version Version

	promptType PromptType
	renderer   PromptRenderer
}

func (p *basePrompt) Type() PromptType {
	return p.promptType
}

func (p *basePrompt) Name() string {
	return p.name
}

func (p *basePrompt) Version() Version {
	return p.version
}

type PromptText struct {
	basePrompt
	content string
}

func NewPromptText(name string, version Version, content string, opts ...PromptOption) *PromptText {
	var cfg promptConfig
	cfg.renderer = defaultRenderer
	for _, opt := range opts {
		opt(&cfg)
	}

	return &PromptText{
		basePrompt: basePrompt{
			name:       name,
			version:    version,
			promptType: PromptTypeText,
			renderer:   cfg.renderer,
		},
		content: content,
	}
}

func (p *PromptText) RawTextTemplate() string {
	return p.content
}

func (p *PromptText) Render(parameters map[string]any) (string, error) {
	return p.renderer.Render(p.content, parameters)
}

type PromptChatMessage struct {
	Role    PromptRole
	Content string
}

type PromptChat struct {
	basePrompt
	messages []PromptChatMessage
}

func NewPromptChat(name string, version Version, messages []PromptChatMessage, opts ...PromptOption) *PromptChat {
	var cfg promptConfig
	cfg.renderer = defaultRenderer
	for _, opt := range opts {
		opt(&cfg)
	}

	return &PromptChat{
		basePrompt: basePrompt{
			name:       name,
			version:    version,
			promptType: PromptTypeChat,
			renderer:   cfg.renderer,
		},
		messages: messages,
	}
}

func (p *PromptChat) RawMessagesTemplate() []PromptChatMessage {
	return p.messages
}

func (p *PromptChat) Render(parms map[string]any) ([]PromptChatMessage, error) {
	var rendered []PromptChatMessage
	for _, msg := range p.messages {
		renderedMsg, err := p.renderer.Render(msg.Content, parms)
		if err != nil {
			return nil, err
		}
		rendered = append(rendered, PromptChatMessage{Role: msg.Role, Content: renderedMsg})
	}
	return rendered, nil
}
