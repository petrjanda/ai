package workflows

import (
	"context"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
)

// TypedWrapper expands on StructuredTask by providing a typed result.
// It is a wrapper and due to typed output breaks Task interface.
// It's a convenience wrapper that could be used as part of wider workflows,
// exposing internal task that implements Task interface that can be used
// in evals and other areas of code that demand Task interface.

type Work[I any, O any] interface {
	Invoke(ctx context.Context, in *I) (*O, error)
}

type Func[I any, O any] = func(ctx context.Context, llm ai.LLM, in *I) (*O, error)

type FunctionWork[I any, O any] struct {
	name string
	fn   Func[I, O]
}

func NewFunctionWork[I any, O any](name string, fn Func[I, O]) *FunctionWork[I, O] {
	return &FunctionWork[I, O]{
		name: name,
		fn:   fn,
	}
}

func (t *FunctionWork[I, O]) Name() string {
	return t.name
}

func (t *FunctionWork[I, O]) Invoke(ctx context.Context, llm ai.LLM, in *I) (*O, error) {
	if response, ok := loadWork[O](ctx, t.name); ok {
		return response, nil
	}

	response, err := t.fn(ctx, llm, in)
	if err != nil {
		return nil, err
	}

	return saveWork(ctx, t.name, response)
}
