package structured

import (
	"context"
	"encoding/json"
	"time"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries   int
	RetryDelay   time.Duration
	RetryBackoff float64
	ModelId      ai.ModelId
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   0,
		RetryDelay:   100 * time.Millisecond,
		RetryBackoff: 2.0,
		ModelId:      ai.Claude4Sonnet,
	}
}

func NewRetryConfig(modelId ai.ModelId, maxRetries int, retryDelay time.Duration, retryBackoff float64) *RetryConfig {
	return &RetryConfig{
		ModelId:      modelId,
		MaxRetries:   maxRetries,
		RetryDelay:   retryDelay,
		RetryBackoff: retryBackoff,
	}
}

// RetryableOperation represents an operation that can be retried
type RetryableOperation[T any] interface {
	// Execute performs the operation and returns the result or error
	Execute(ctx context.Context, attempt int) (T, error)

	// GetCorrectionPrompt returns a prompt for the LLM to correct the operation
	GetCorrectionPrompt(ctx context.Context, lastError error) string

	// UpdateWithCorrection updates the operation with corrected parameters from the LLM
	UpdateWithCorrection(ctx context.Context, correction json.RawMessage) error

	// GetSchema returns the JSON schema for validation/correction
	GetSchema() json.RawMessage
}
