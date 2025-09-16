package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/getsynq/ai/pkg/ai"
	"github.com/getsynq/ai/pkg/ai/tools"
)

// ToolCallOperation implements RetryableOperation for tool calls
type ToolCallOperation struct {
	toolCall   *tools.ToolCall
	targetTool tools.Tool
	events     ai.AgentEvents
}

// NewToolCallOperation creates a new tool call operation
func NewToolCallOperation(toolCall *tools.ToolCall, targetTool tools.Tool, events ai.AgentEvents) *ToolCallOperation {
	return &ToolCallOperation{
		toolCall:   toolCall,
		targetTool: targetTool,
		events:     events,
	}
}

// Execute performs the tool call
func (op *ToolCallOperation) Execute(ctx context.Context, attempt int) (ai.Message, error) {
	result, err := op.targetTool.Execute(ctx, op.toolCall.Args)
	if err != nil {
		op.events.OnToolError(ctx, op.toolCall, attempt, err)
		return nil, err
	}

	op.events.OnToolResult(ctx, op.toolCall, result)
	return ai.NewToolResultMessage(op.toolCall, result), nil
}

// GeoprrectionPrompt returns a prompt for correcting the tool call
func (op *ToolCallOperation) GetCorrectionPrompt(ctx context.Context, lastError error) string {
	// Use the same prompt format as the agent
	correctionPromptFormat := `Tool call to '%s' failed with error: %s
Failed parameters: %s
Use 'formatter' tool to generate corrected parameters that match the tool's input schema above.`

	return fmt.Sprintf(
		correctionPromptFormat,
		op.toolCall.Name,
		lastError.Error(),
		prettyJSON(op.toolCall.Args),
	)
}

// UpdateWithCorrection updates the tool call with corrected parameters
func (op *ToolCallOperation) UpdateWithCorrection(ctx context.Context, correction json.RawMessage) error {
	op.toolCall = &tools.ToolCall{
		ID:   op.toolCall.ID,
		Name: op.toolCall.Name,
		Args: correction,
	}
	return nil
}

// GetSchema returns the tool's input schema
func (op *ToolCallOperation) GetSchema() json.RawMessage {
	return op.targetTool.InputSchemaRaw()
}

// prettyJSON formats JSON data for logging
func prettyJSON(data json.RawMessage) string {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, data, "", "  ")
	if err != nil {
		return string(data)
	}
	return prettyJSON.String()
}
