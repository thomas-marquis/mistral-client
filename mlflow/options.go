package mlflow

type promptConfig struct {
	renderer PromptRenderer
}

type PromptOption func(*promptConfig)

func WithCustomPromptRenderer(renderer PromptRenderer) PromptOption {
	return func(cfg *promptConfig) {
		cfg.renderer = renderer
	}
}

func WithSimpleFormatRenderer() PromptOption {
	return func(cfg *promptConfig) {
		cfg.renderer = &simpleFormatRenderer{}
	}
}

func WithRawRenderer() PromptOption {
	return func(cfg *promptConfig) {
		cfg.renderer = &rawRenderer{}
	}
}
