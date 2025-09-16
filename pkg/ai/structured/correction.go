package structured

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
)

type Corrector struct {
	modelId ai.ModelId
	schema  json.RawMessage
	system  string
}

func NewCorrector(modelId ai.ModelId, schema json.RawMessage, system string) *Corrector {
	return &Corrector{
		modelId: modelId,
		schema:  schema,
	}
}

func (c *Corrector) Execute(ctx context.Context, llm ai.LLM, history ai.History) (json.RawMessage, error) {
	retryRequest := ai.NewLLMRequest(
		ai.WithSystem(c.system),
		ai.WithHistory(history),
		ai.WithModel(c.modelId),
		ai.WithTemperature(0.1),
	)

	structuredLLM := NewLLM(c.schema, llm)
	retryResponse, retryErr := structuredLLM.Invoke(ctx, retryRequest)
	if retryErr != nil {
		return nil, fmt.Errorf("failed to get corrected parameters: %w", retryErr)
	}

	if len(retryResponse.ToolCalls()) > 0 {
		toolCall := retryResponse.ToolCalls()[0]
		return toolCall.Args, nil
	}

	return nil, fmt.Errorf("structured LLM did not provide corrected parameters")
}
