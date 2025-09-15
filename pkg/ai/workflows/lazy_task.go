package workflows

import (
	"context"

	"github.com/getsynq/ai/pkg/ai"
)

// LazyTask is a task that defers execution until it's invoked.
// It simply executes a callback function when invoked, without wrapping another task.
// This is useful for cases where you want to delay expensive operations like
// data preloading until the task is actually executed.
type LazyTask struct {
	Name_    string
	callback LazyTaskCallback
}

// LazyTaskCallback is a function that performs the actual work when the task is invoked.
// It receives the context, LLM, and history and should return an LLMResponse and any error.
type LazyTaskCallback func(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error)

// NewLazyTask creates a new lazy task with a callback that will be executed
// when the task is invoked.
func NewLazyTask(name string, callback LazyTaskCallback) *LazyTask {
	return &LazyTask{
		Name_:    name,
		callback: callback,
	}
}

func (t *LazyTask) Name() string {
	return t.Name_
}

func (t *LazyTask) Clone() Task {
	return &LazyTask{
		Name_:    t.Name_,
		callback: t.callback,
	}
}

func (t *LazyTask) Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
	// Simply execute the callback
	response, err := t.callback(ctx, llm, history)
	if err != nil {
		return nil, err
	}

	response.Messages = history.Append(response.Messages...)

	return response, nil
}

func (t *LazyTask) WithName(name string) Task {
	new := t.Clone().(*LazyTask)
	new.Name_ = name
	return new
}

func (t *LazyTask) WithRequestOpts(opts ...ai.LLMRequestOpts) Task {
	// LazyTask doesn't support request options since it's just a callback executor
	// Return a clone without applying the options
	return t.Clone()
}

func (t *LazyTask) Then(task Task) Task {
	return NewChainTask(t, task, false)
}

func (t *LazyTask) Pipe(task Task) Task {
	return NewChainTask(t, task, true)
}
