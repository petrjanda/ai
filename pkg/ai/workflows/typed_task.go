package workflows

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/structured"
)

// TypedWrapper expands on StructuredTask by providing a typed result.
// It is a wrapper and due to typed output breaks Task interface.
// It's a convenience wrapper that could be used as part of wider workflows,
// exposing internal task that implements Task interface that can be used
// in evals and other areas of code that demand Task interface.

type typed[T any] struct {
	Inner Task
}

func NewTypedTask[T any](name string, request *ai.LLMRequest, opts ...structured.LLMOpts) *typed[T] {
	task := NewStructuredTask[T](name, request, opts...)
	return &typed[T]{
		Inner: task,
	}
}

func Typed[T any](task Task) *typed[T] {
	return &typed[T]{
		Inner: task,
	}
}

func (t *typed[T]) Name() string {
	return t.Inner.Name()
}

func (t *typed[T]) InvokeTyped(ctx context.Context, llm ai.LLM, history ai.History) (*T, error) {
	response, err := t.Inner.Invoke(ctx, llm, history)
	if err != nil {
		return new(T), err
	}

	lastMessage := response.LastMessageAsText()
	if lastMessage == nil {
		return new(T), fmt.Errorf("last message is not a text message")
	}

	var result T
	err = json.Unmarshal([]byte(lastMessage.Content), &result)
	if err != nil {
		return new(T), err
	}

	return &result, nil
}

func (t *typed[T]) Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
	return t.Inner.Invoke(ctx, llm, history)
}

func (t *typed[T]) Clone() Task {
	return &typed[T]{
		Inner: t.Inner.Clone(),
	}
}

func (t *typed[T]) WithRequestOpts(opts ...ai.LLMRequestOpts) Task {
	return &typed[T]{
		Inner: t.Inner.WithRequestOpts(opts...),
	}
}

func (t *typed[T]) WithName(name string) Task {
	return &typed[T]{
		Inner: t.Inner.WithName(name),
	}
}

func (t *typed[T]) Pipe(task Task) Task {
	return NewChainTask(t, task, false)
}

func (t *typed[T]) Then(task Task) Task {
	return NewChainTask(t, task, true)
}
