package workflows

import (
	"context"
	"encoding/json"

	"github.com/getsynq/ai/pkg/ai"
	"github.com/getsynq/ai/pkg/ai/prompts"
	"github.com/getsynq/ai/pkg/ai/tools"
)

type ToolTask struct {
	Name_  string
	Tool   tools.Tool
	Args   json.RawMessage
	Format HistoryFunc
}

type HistoryFunc func(msg json.RawMessage, title string) ai.History

type ToolTaskOpts func(task *ToolTask)

func ToolTaskWithFormat(format HistoryFunc) ToolTaskOpts {
	return func(task *ToolTask) {
		task.Format = format
	}
}

func NewToolTask(name string, tool tools.Tool, args json.RawMessage, opts ...ToolTaskOpts) Task {
	task := &ToolTask{
		Name_:  name,
		Tool:   tool,
		Args:   args,
		Format: toolTaskToHistory,
	}

	for _, opt := range opts {
		opt(task)
	}

	return NewLazyTask(task.Name_, func(ctx context.Context, _ ai.LLM, _ ai.History) (*ai.LLMResponse, error) {
		result, err := task.Tool.Execute(ctx, task.Args)
		if err != nil {
			return nil, err
		}
		return ai.NewLLMResponse(toolTaskToHistory(result, task.Name_)...), nil
	})
}

func toolTaskToHistory(msg json.RawMessage, title string) ai.History {
	return ai.NewHistory(
		prompts.
			NewPromptBuilder().
			AddBlock(string(msg), prompts.WithTitle(title)).
			BuildUserMessage(),
	)
}
