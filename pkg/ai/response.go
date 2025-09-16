package ai

import "github.com/getsynq/cloud/ai-data-sre/pkg/ai/tools"

type LLMResponse struct {
	Messages History   `json:"messages"`
	Usage    *LLMUsage `json:"usage"`
}

func NewLLMResponse(messages ...Message) *LLMResponse {
	return &LLMResponse{
		Messages: messages,
		Usage:    NewLLMUsage(0, 0, 0),
	}
}

func (r *LLMResponse) AddMessage(message Message) {
	r.Messages = r.Messages.Append(message)
}

func (r *LLMResponse) AddToolCall(functionCall *tools.ToolCall) {
	r.Messages = r.Messages.Append(NewToolCallMessage(functionCall))
}

func (r *LLMResponse) Clone() *LLMResponse {
	messages := make(History, len(r.Messages))
	copy(messages, r.Messages)

	return &LLMResponse{
		Messages: messages,
		Usage:    r.Usage,
	}
}

func (r *LLMResponse) ToolCalls() []*tools.ToolCall {
	var toolCalls []*tools.ToolCall
	for _, msg := range r.Messages {
		if msg.Kind() == MessageKindToolCall {
			toolCalls = append(toolCalls, msg.(*ToolCallMessage).ToolCall)
		}
	}

	return toolCalls
}

func (r *LLMResponse) LastMessageAsText() *TextMessage {
	if len(r.Messages) == 0 {
		return nil
	}

	text, ok := r.Messages[len(r.Messages)-1].(*TextMessage)
	if !ok {
		return nil
	}

	return text
}

func (r *LLMResponse) SetUsage(usage *LLMUsage) {
	r.Usage = usage
}

func (r *LLMResponse) AddUsage(usage *LLMUsage) {
	r.Usage.Add(usage)
}
