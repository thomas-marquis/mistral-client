package mlflow

import "context"

type promptConfig struct {
	renderer PromptRenderer
	ctx      context.Context
	headers  map[string]string
}

type PromptOption func(*promptConfig)

// WithCustomPromptRenderer allows you to set your own PromptRenderer implementation
func WithCustomPromptRenderer(renderer PromptRenderer) PromptOption {
	return func(cfg *promptConfig) {
		cfg.renderer = renderer
	}
}

// WithSimpleFormatRenderer set a simple format renderer.
// It simply replaces double braces placeholders with corresponding parameter values.
// Examples:
//
//	"{{name}}" => "John"
//	"{{ name }}" => "John"
//
// This renderer doesn't support nested parameters.
func WithSimpleFormatRenderer() PromptOption {
	return func(cfg *promptConfig) {
		cfg.renderer = &simpleFormatRenderer{}
	}
}

// WithRawRenderer renders prompt content as is.
func WithRawRenderer() PromptOption {
	return func(cfg *promptConfig) {
		cfg.renderer = &rawRenderer{}
	}
}

// WithGoTemplateRenderer renders prompt content using Go template engine with sprig functions.
// See https://pkg.go.dev/text/template and https://masterminds.github.io/sprig/ for more details.
func WithGoTemplateRenderer() PromptOption {
	return func(cfg *promptConfig) {
		cfg.renderer = &goTemplateRenderer{}
	}
}

// WithHeaders sets custom headers for HTTP requests.
func WithHeaders(headers map[string]string) PromptOption {
	return func(cfg *promptConfig) {
		cfg.headers = headers
	}
}
