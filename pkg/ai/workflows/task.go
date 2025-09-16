package workflows

import (
	"context"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
)

type Task interface {
	Name() string
	Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error)
	Clone() Task

	// Concatenates tasks and passes through the entire history
	Then(task Task) Task

	// Concatenates tasks and passes through the last message in history
	Pipe(task Task) Task

	WithName(name string) Task
	WithRequestOpts(opts ...ai.LLMRequestOpts) Task
}
