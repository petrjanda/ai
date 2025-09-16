package workflows

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/structured"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/tools"
)

// StructuredTask is a task that returns a structured response.
// It wraps invocation LLM with StructuredLLM to enforce output schema,
// but it still returns LLMResponse with all its details.

type StructuredTask struct {
	Name_ string

	Request *ai.LLMRequest
	Opts    []structured.LLMOpts

	schema json.RawMessage
}

type StructuredTaskOpts = func(*StructuredTask)

func StructuredTaskWithRequest(request *ai.LLMRequest) StructuredTaskOpts {
	return func(t *StructuredTask) {
		t.Request = request
	}
}

// NewStructuredTask creates a new structured task with a pre-generated schema
func NewStructuredTask[T any](name string, request *ai.LLMRequest, opts ...structured.LLMOpts) *StructuredTask {
	schema := tools.DefaultSchemaGenerator.MustGenerate(new(T))

	return &StructuredTask{
		Name_:   name,
		Request: request,
		Opts:    opts,
		schema:  schema,
	}
}

func (t *StructuredTask) Name() string {
	return t.Name_
}

func (t *StructuredTask) Clone() Task {
	return &StructuredTask{
		Name_:   t.Name_,
		Request: t.Request.Clone(),
	}
}

func (t *StructuredTask) Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {

	// Create structured LLM with the pre-generated schema
	structuredLLM := structured.NewLLM(t.schema, llm, t.Opts...)

	// Invoke the structured LLM
	response, err := structuredLLM.Invoke(ctx, t.Request.Clone(ai.WithAddedHistory(history)))
	if err != nil {
		return nil, err
	}

	// Validate that the last message is text
	lastMessage := response.LastMessageAsText()
	if lastMessage == nil {
		return nil, fmt.Errorf("last message is not a text message")
	}

	return response, nil
}

func (t *StructuredTask) With(opt StructuredTaskOpts) *StructuredTask {
	opt(t)
	return t
}

func (t *StructuredTask) WithRequestOpts(opts ...ai.LLMRequestOpts) Task {
	new := t.Clone().(*StructuredTask)
	for _, opt := range opts {
		opt(new.Request)
	}
	return new
}

func (t *StructuredTask) WithName(name string) Task {
	new := t.Clone().(*StructuredTask)
	new.Name_ = name
	return new
}

func (t *StructuredTask) Then(task Task) Task {
	return NewChainTask(t, task, false)
}

func (t *StructuredTask) Pipe(task Task) Task {
	return NewChainTask(t, task, true)
}
