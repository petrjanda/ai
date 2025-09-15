package ai

import (
	"encoding/json"

	"github.com/getsynq/ai/pkg/ai/tools"
)

// Message represents a message in the conversation
type Message interface {
	Kind() MessageKind
	Role() MessageRole
}

// MessageKind represents the type of message
type MessageKind string

const (
	MessageKindText       MessageKind = "text"
	MessageKindToolCall   MessageKind = "tool_call"
	MessageKindToolResult MessageKind = "tool_result"
)

// MessageRole represents the role of the message sender
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
	MessageRoleTool      MessageRole = "tool"
)

type TextMessage struct {
	Content string      `json:"content"`
	Role_   MessageRole `json:"role"`
}

func (m *TextMessage) Kind() MessageKind {
	return MessageKindText
}

func (m *TextMessage) Role() MessageRole {
	return m.Role_
}

func NewUserMessage(content string) *TextMessage {
	return &TextMessage{
		Content: content,
		Role_:   MessageRoleUser,
	}
}

func NewAssistantMessage(content string) *TextMessage {
	return &TextMessage{
		Content: content,
		Role_:   MessageRoleAssistant,
	}
}

func NewSystemMessage(content string) *TextMessage {
	return &TextMessage{
		Content: content,
		Role_:   MessageRoleSystem,
	}
}

type ToolCallMessage struct {
	ToolCall *tools.ToolCall `json:"tool_call"`
}

func NewToolCallMessage(toolCall *tools.ToolCall) *ToolCallMessage {
	return &ToolCallMessage{
		ToolCall: toolCall,
	}
}

func (m *ToolCallMessage) Kind() MessageKind {
	return MessageKindToolCall
}

func (m *ToolCallMessage) Role() MessageRole {
	return MessageRoleAssistant
}

// ToolResultMessage represents the result of a tool execution
type ToolResultMessage struct {
	ToolCall *tools.ToolCall `json:"tool_call"`
	Result   json.RawMessage `json:"result"`
	Error    string          `json:"error,omitempty"`
}

func NewToolResultMessage(toolCall *tools.ToolCall, result json.RawMessage) *ToolResultMessage {
	return &ToolResultMessage{
		ToolCall: toolCall,
		Result:   result,
	}
}

func NewToolResultErrorMessage(toolCall *tools.ToolCall, err_ string) *ToolResultMessage {
	return &ToolResultMessage{
		ToolCall: toolCall,
		Error:    err_,
	}
}

func (m *ToolResultMessage) Kind() MessageKind {
	return MessageKindToolResult
}

func (m *ToolResultMessage) Role() MessageRole {
	return MessageRoleTool
}

// ToolErrorMessage represents an error that occurred during tool execution
type ToolErrorMessage struct {
	ToolCall *tools.ToolCall
	Error    string
}

func NewToolErrorMessage(toolCall *tools.ToolCall, error string) *ToolErrorMessage {
	return &ToolErrorMessage{
		ToolCall: toolCall,
		Error:    error,
	}
}

func (m *ToolErrorMessage) Kind() MessageKind {
	return MessageKindToolResult
}

func (m *ToolErrorMessage) Role() MessageRole {
	return MessageRoleAssistant
}
