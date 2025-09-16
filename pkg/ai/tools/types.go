package tools

import "encoding/json"

// ToolCall represents a call to a tool
type ToolCall struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

func NewToolCall(id, name string, args json.RawMessage) *ToolCall {
	return &ToolCall{
		ID:   id,
		Name: name,
		Args: args,
	}
}
