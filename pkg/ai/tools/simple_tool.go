package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// SimpleTool is a generic tool implementation that handles JSON marshalling/unmarshalling
// automatically for input type I and output type O
type SimpleTool[I, O any] struct {
	name        string
	description string
	runner      func(ctx context.Context, input *I) (*O, error)
}

// NewSimpleTool creates a new generic tool with the given name, description, and runner function
func NewSimpleTool[I, O any](name, description string, runner func(ctx context.Context, input *I) (*O, error)) *SimpleTool[I, O] {
	return &SimpleTool[I, O]{
		name:        name,
		description: description,
		runner:      runner,
	}
}

// Name returns the name of the tool
func (g *SimpleTool[I, O]) Name() string {
	return g.name
}

// Description returns the description of the tool
func (g *SimpleTool[I, O]) Description() string {
	return g.description
}

// InputSchemaRaw returns the JSON schema for the tool's input type I
func (g *SimpleTool[I, O]) InputSchemaRaw() json.RawMessage {
	return DefaultSchemaGenerator.MustGenerate(new(I))
}

// OutputSchemaRaw returns the JSON schema for the tool's output type O
func (g *SimpleTool[I, O]) OutputSchemaRaw() json.RawMessage {
	return DefaultSchemaGenerator.MustGenerate(new(O))
}

// Run executes the tool with the given arguments, automatically handling JSON marshalling/unmarshalling
func (g *SimpleTool[I, O]) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	// Unmarshal the input arguments to type I
	var input I
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	// Run the tool with the typed input
	output, err := g.runner(ctx, &input)
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
