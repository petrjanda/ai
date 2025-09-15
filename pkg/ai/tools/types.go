package tools

import "encoding/json"

// ToolCall represents a call to a tool
type ToolCall struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}
