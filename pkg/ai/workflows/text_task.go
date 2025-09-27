package workflows

import (
	"context"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
)

// TextTask is a simple task that returns a text response.

type TextTask struct {
	Name_   string
	Request *ai.LLMRequest
}

func NewTask(name string, request *ai.LLMRequest) Task {
	return &TextTask{
		Name_:   name,
		Request: request,
	}
}

func (t *TextTask) Name() string {
	return t.Name_
}

func (t *TextTask) Clone() Task {
	return NewTask(t.Name_, t.Request.Clone())
}

func (t *TextTask) Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
	if response, ok := loadTask(ctx, t.Name_); ok {
		return response, nil
	}

	response, err := llm.Invoke(ctx, t.Request.Clone(ai.WithHistory(history)))
	if err != nil {
		return nil, err
	}

	return saveTask(ctx, t.Name_, response)
}

func (t *TextTask) WithRequestOpts(opts ...ai.LLMRequestOpts) Task {
	new := t.Clone().(*TextTask)
	for _, opt := range opts {
		opt(new.Request)
	}

	return new
}

func (t *TextTask) WithName(name string) Task {
	new := t.Clone().(*TextTask)
	new.Name_ = name
	return new
}

func (t *TextTask) Then(task Task) Task {
	return NewChainTask(t, task, false)
}

func (t *TextTask) Pipe(task Task) Task {
	return NewChainTask(t, task, true)
}
