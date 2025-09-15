package openai

import (
	"fmt"

	"github.com/getsynq/ai/pkg/ai/tools"
	"github.com/openai/openai-go/v2"
)

// convertToolUsage converts our ToolUsage interface to OpenAI's tool choice format
// Returns nil when no specific tool choice is needed (auto/default behavior)
func convertToolUsage(toolUsage tools.ToolUsage, tools_ tools.Toolbox) (*openai.ChatCompletionToolChoiceOptionUnionParam, error) {
	switch toolUsage.Type() {
	default:
		return nil, nil

	case tools.ToolUsageAuto:
		return nil, nil

	case tools.ToolUsageForced:
		if forced, ok := toolUsage.(*tools.ForcedToolUsage); ok {
			tool, err := tools_.FindTool(forced.ToolName)
			if err != nil {
				return nil, fmt.Errorf("forced tool %s not available", forced.ToolName)
			}

			toolChoice := openai.ToolChoiceOptionFunctionToolChoice(openai.ChatCompletionNamedToolChoiceFunctionParam{
				Name: tool.Name(),
			})
			return &toolChoice, nil
		}

		return nil, nil
	}
}
