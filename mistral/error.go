package mistral

import (
	"fmt"
	"strings"
)

type ApiError struct {
	Code    int
	Content map[string]any
}

func (e *ApiError) Error() string {
	msg := strings.Builder{}
	if e.Code > 0 {
		fmt.Fprintf(&msg, "[%d] ", e.Code)
	}

	if e.Content == nil {
		return strings.TrimSpace(msg.String())
	}

	// Extract error type
	errType, _ := e.Content["type"].(string)

	var details []string

	// Check "message" field which can be a string or an object containing details
	if rawMsg, ok := e.Content["message"]; ok {
		if m, ok := rawMsg.(string); ok {
			details = append(details, m)
		} else if m, ok := rawMsg.(map[string]any); ok {
			if rawDetail, ok := m["detail"]; ok {
				details = append(details, extractApiErrorDetails(rawDetail)...)
			}
		}
	} else if detail, ok := e.Content["detail"].(string); ok {
		// Case for 401 Unauthorized: {"detail": "Unauthorized"}
		details = append(details, detail)
	}

	if errType != "" {
		msg.WriteString(errType)
		if len(details) > 0 {
			msg.WriteString(": ")
		}
	}

	if len(details) > 0 {
		msg.WriteString(strings.Join(details, "; "))
	}

	return strings.TrimSpace(msg.String())
}

func extractApiErrorDetails(raw any) []string {
	var details []string
	switch v := raw.(type) {
	case []any:
		for _, d := range v {
			if dm, ok := d.(map[string]any); ok {
				details = append(details, formatApiErrorDetail(dm))
			}
		}
	case []map[string]any:
		for _, dm := range v {
			details = append(details, formatApiErrorDetail(dm))
		}
	}
	return details
}

func formatApiErrorDetail(d map[string]any) string {
	var sb strings.Builder

	t, _ := d["type"].(string)
	m, _ := d["msg"].(string)

	if t != "" {
		sb.WriteString(t)
		sb.WriteString(": ")
	}
	sb.WriteString(m)

	if locRaw, ok := d["loc"]; ok {
		var locParts []string
		switch lv := locRaw.(type) {
		case []any:
			for _, l := range lv {
				if s, ok := l.(string); ok {
					locParts = append(locParts, s)
				}
			}
		case []string:
			locParts = lv
		}
		if len(locParts) > 0 {
			sb.WriteString(" (")
			sb.WriteString(strings.Join(locParts, "."))
			sb.WriteString(")")
		}
	}

	return sb.String()
}
