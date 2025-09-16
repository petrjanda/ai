package structured

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"

	"github.com/pkg/errors"
)

// StructuredOutputOperation implements RetryableOperation for structured output generation
type StructuredOutputOperation struct {
	request         *ai.LLMRequest
	llm             ai.LLM
	formatter       StructuredLLM
	events          ai.AgentEvents
	originalRequest *ai.LLMRequest
}

// NewStructuredOutputOperation creates a new structured output operation
func NewStructuredOutputOperation(request *ai.LLMRequest, llm ai.LLM, formatter StructuredLLM, events ai.AgentEvents) *StructuredOutputOperation {
	return &StructuredOutputOperation{
		request:         request,
		llm:             llm,
		formatter:       formatter,
		events:          events,
		originalRequest: request.Clone(), // Keep a copy of the original request
	}
}

// Execute performs the structured output generation
func (op *StructuredOutputOperation) Execute(ctx context.Context, attempt int) (*ai.LLMResponse, error) {

	// Delegate to the underlying LLM
	response, err := op.llm.Invoke(ctx, op.request)
	if err != nil {
		return nil, errors.Wrap(err, "underlying LLM invocation failed")
	}

	// Verify the LLM followed forced tool usage
	toolCalls := response.ToolCalls()
	if len(toolCalls) == 0 {
		return nil, errors.New("no tool call found in response - LLM did not follow forced tool usage")
	}

	// Verify the format of produced structured output
	toolCall := toolCalls[0]
	result, err := op.formatter.Execute(ctx, toolCall.Args)
	if err != nil {
		op.events.OnToolError(ctx, toolCall, attempt, err)
		return nil, errors.Wrap(err, "invalid structured output")
	}

	// Verify it can be marshalled
	if _, err := json.Marshal(result); err != nil {
		op.events.OnToolError(ctx, toolCall, attempt, err)
		return nil, errors.Wrap(err, "structured output marshalling failed")
	}

	return response, nil
}

// GetCorrectionPrompt returns a prompt for correcting the structured output
func (op *StructuredOutputOperation) GetCorrectionPrompt(ctx context.Context, lastError error) string {
	// Create a detailed prompt for structured output correction
	correctionPromptFormat := `Structured output generation failed with error: %s

The LLM was asked to generate structured output that matches the provided schema, but the generation failed.

Please provide corrected structured output that matches the schema requirements.

Original request context:
%s

Please ensure the output:
1. Matches the exact schema structure
2. Contains all required fields
3. Has correct data types for each field
4. Follows any format constraints specified in the schema`

	// Get a summary of the original request for context
	requestSummary := fmt.Sprintf("Messages: %d, Tools: %d", len(op.originalRequest.History), len(op.originalRequest.Tools))

	return fmt.Sprintf(
		correctionPromptFormat,
		lastError.Error(),
		requestSummary,
	)
}

func (op *StructuredOutputOperation) UpdateWithCorrection(ctx context.Context, correction json.RawMessage) error {
	return nil
}

// GetSchema returns the structured output schema
func (op *StructuredOutputOperation) GetSchema() json.RawMessage {
	return op.formatter.InputSchemaRaw()
}
