package mlflow

import "fmt"

type PromptRenderer interface {
	Render(raw string, params map[string]any) (string, error)
}

var (
	_ PromptRenderer = (*simpleFormatRenderer)(nil)
	_ PromptRenderer = (*rawRenderer)(nil)

	defaultRenderer = &simpleFormatRenderer{}
)

type simpleFormatRenderer struct{}

func (r *simpleFormatRenderer) Render(raw string, params map[string]any) (string, error) {
	if params == nil {
		return raw, nil
	}
	// replace {{...}} with coresponding parameters.
	// no nested map allowed
	return fmt.Sprintf(raw, params), nil
}

type rawRenderer struct{}

func (r *rawRenderer) Render(raw string, _ map[string]any) (string, error) {
	return raw, nil
}
