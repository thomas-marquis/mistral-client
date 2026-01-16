package mistral

type UsageInfo struct {
	CompletionTokens   int `json:"completion_tokens,omitempty"`
	PromptAudioSeconds int `json:"prompt_audio_seconds,omitempty"`
	PromptTokens       int `json:"prompt_tokens,omitempty"`
	TotalTokens        int `json:"total_tokens,omitempty"`
}
