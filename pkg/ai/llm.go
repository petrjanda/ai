package ai

import (
	"context"
)

// LLM represents a language model that can process requests and generate responses
type LLM interface {
	Invoke(ctx context.Context, request *LLMRequest) (*LLMResponse, error)
}
