package mlflow

type modelVersion struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	CreationTimestamp    int64             `json:"creation_timestamp,omitempty"`
	LastUpdatedTimestamp int64             `json:"last_updated_timestamp,omitempty"`
	UserID               string            `json:"user_id,omitempty"`
	CurrentStage         string            `json:"current_stage,omitempty"`
	Description          string            `json:"description,omitempty"`
	Source               string            `json:"source,omitempty"`
	RunID                string            `json:"run_id,omitempty"`
	Status               string            `json:"status,omitempty"`
	StatusMessage        string            `json:"status_message,omitempty"`
	Tags                 []modelVersionTag `json:"tags,omitempty"`
	RunLink              string            `json:"run_link,omitempty"`
	Aliases              []string          `json:"aliases,omitempty"`
}

type modelVersionTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
