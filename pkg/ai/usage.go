package ai

import (
	"encoding/json"
	"fmt"

	"github.com/getsynq/ai/pkg/ai/tools"
)

type LLMUsageTokens struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

type LLMUsage struct {
	LLMUsageTokens `json:",inline"`
	Turns          int64 `json:"turns"`
	ToolCalls      []*LLMUsageToolCall
}

type LLMUsageToolCall struct {
	Name  string          `json:"name"`
	Args  json.RawMessage `json:"args"`
	Error error           `json:"error"`
}

func NewLLMUsage(promptTokens, completionTokens, totalTokens int64) *LLMUsage {
	return &LLMUsage{
		LLMUsageTokens: LLMUsageTokens{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		},
		Turns: 1,
	}
}

func (u *LLMUsage) Add(other *LLMUsage) {
	u.PromptTokens += other.PromptTokens
	u.CompletionTokens += other.CompletionTokens
	u.TotalTokens += other.TotalTokens
	u.Turns += other.Turns

	for _, toolCall := range other.ToolCalls {
		u.ToolCalls = append(u.ToolCalls, toolCall)
	}
}

func (u *LLMUsage) AddToolCall(toolCall *tools.ToolCall, err error) {
	u.ToolCalls = append(u.ToolCalls, &LLMUsageToolCall{
		Name:  toolCall.Name,
		Args:  toolCall.Args,
		Error: err,
	})
}

func (u *LLMUsage) String() string {
	summary := fmt.Sprintf("prompt: %d, completion: %d, total: %d, tools: %v", u.PromptTokens, u.CompletionTokens, u.TotalTokens, len(u.ToolCalls))
	for _, toolCall := range u.ToolCalls {
		if toolCall.Error != nil {
			summary += fmt.Sprintf("\n  - [ERR] %s, %s, %s", toolCall.Error, toolCall.Name, toolCall.Args)
		} else {
			summary += fmt.Sprintf("\n  - %s, %s", toolCall.Name, toolCall.Args)
		}
	}
	return summary
}
