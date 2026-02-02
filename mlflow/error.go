package mlflow

import "fmt"

// APIError represents an error returned by the MLflow API.
type APIError struct {
	StatusCode int    `json:"-"`
	ErrorCode  string `json:"error_code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.ErrorCode != "" {
		return fmt.Sprintf("mlflow api error [%s]: %s", e.ErrorCode, e.Message)
	}
	return fmt.Sprintf("mlflow api error (status %d): %s", e.StatusCode, e.Message)
}
