package mlflow

import (
	"fmt"
	"regexp"
)

type PromptRenderer interface {
	Render(raw string, params map[string]any) (string, error)
}

var (
	_ PromptRenderer = (*simpleFormatRenderer)(nil)
	_ PromptRenderer = (*rawRenderer)(nil)

	defaultRenderer = &simpleFormatRenderer{}
)

type simpleFormatRenderer struct{}

var rePlaceholder = regexp.MustCompile(`{{\s*([\w.]+)\s*}}`)

func (r *simpleFormatRenderer) Render(raw string, params map[string]any) (string, error) {
	if params == nil {
		return raw, nil
	}

	for key, val := range params {
		if _, ok := val.(map[string]any); ok {
			return "", fmt.Errorf("nested parameters are not supported: %s", key)
		}
	}

	var renderErr error
	result := rePlaceholder.ReplaceAllStringFunc(raw, func(match string) string {
		if renderErr != nil {
			return match
		}

		key := rePlaceholder.FindStringSubmatch(match)[1]

		if val, ok := params[key]; ok {
			return fmt.Sprintf("%v", val)
		}

		renderErr = fmt.Errorf("missing parameter: %s", key)
		return match
	})

	if renderErr != nil {
		return "", renderErr
	}

	return result, nil
}

type rawRenderer struct{}

func (r *rawRenderer) Render(raw string, _ map[string]any) (string, error) {
	return raw, nil
}
