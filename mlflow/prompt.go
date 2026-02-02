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
	String() string
	Render(parameters map[string]any) (string, error)
}

var (
	_ Prompt = (*PromptText)(nil)
	_ Prompt = (*PromptChat)(nil)
)

type basePrompt struct {
	name    string
	version Version

	content    string
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

func (p *basePrompt) String() string {
	return p.content
}

func (p *basePrompt) Render(parameters map[string]any) (string, error) {
	return p.renderer.Render(p.content, parameters)
}

type PromptText struct {
	basePrompt
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
			content:    content,
			renderer:   cfg.renderer,
		},
	}
}

type PromptChat struct {
	basePrompt
	Role PromptRole
}

func NewPromptChat(name string, version Version, content string, role PromptRole, opts ...PromptOption) *PromptChat {
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
			content:    content,
			renderer:   cfg.renderer,
		},
		Role: role,
	}
}

type PromptTemplate interface {
	String() string
	Render()
}
