package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// GoTool is a generic tool interface that can be implemented to provide tools to LLM
// agents.

type GoTool[I, O any] interface {
	Name() string

	Description() string

	Run(ctx context.Context, input *I) (*O, error)
}

// Adapter takes go tool interface and implements Tool-compatible interface so go tool
// can be used as a tool in the agent.

type Adapter[I, O any] struct {
	tool GoTool[I, O]
}

func NewAdapter[I, O any](tool GoTool[I, O]) *Adapter[I, O] {
	return &Adapter[I, O]{
		tool: tool,
	}
}

func (a *Adapter[I, O]) Name() string {
	return a.tool.Name()
}

func (a *Adapter[I, O]) Description() string {
	return a.tool.Description()
}

func (a *Adapter[I, O]) InputSchemaRaw() json.RawMessage {
	return DefaultSchemaGenerator.MustGenerate(new(I))
}

func (a *Adapter[I, O]) OutputSchemaRaw() json.RawMessage {
	return DefaultSchemaGenerator.MustGenerate(new(O))
}

func (a *Adapter[I, O]) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	// Unmarshal the input arguments to type I
	var input I
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	// Run the tool with the typed input
	output, err := a.tool.Run(ctx, &input)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	// Marshal the output to JSON
	result, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return json.RawMessage(result), nil
}
