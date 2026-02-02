package mlflow

type PromptRenderer interface {
	Render() string
}

type promptRendererPythonFormat struct {
	prompt *Prompt
}
