package workflows

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
)

// TypedWrapper expands on StructuredTask by providing a typed result.
// It is a wrapper and due to typed output breaks Task interface.
// It's a convenience wrapper that could be used as part of wider workflows,
// exposing internal task that implements Task interface that can be used
// in evals and other areas of code that demand Task interface.

type TypedWrapper[T any] struct {
	Inner *StructuredTask
}

func NewTypedTask[T any](name string, request *ai.LLMRequest) (*TypedWrapper[T], error) {
	task := NewStructuredTask[T](name, request)
	return &TypedWrapper[T]{
		Inner: task,
	}, nil
}

func NewTypedWrapper[T any](task *StructuredTask) *TypedWrapper[T] {
	return &TypedWrapper[T]{
		Inner: task,
	}
}

func (t *TypedWrapper[T]) Name() string {
	return t.Inner.Name()
}

func (t *TypedWrapper[T]) Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*T, error) {
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
