package structured

import (
	"context"
	"fmt"
	"time"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
)

type Retryable[T any] interface {
	Retry(ctx context.Context, attempt int) (T, error)
	OnFailure(ctx context.Context, attempt int, err error) error
}

type Retrier[T any] struct {
	config             *RetryConfig
	retriableOperation Retryable[T]
}

func NewRetrier[T any](config *RetryConfig, retriableOperation Retryable[T]) *Retrier[T] {
	return &Retrier[T]{config: config, retriableOperation: retriableOperation}
}

func (r *Retrier[T]) Execute(ctx context.Context, llm ai.LLM) (T, error) {
	var zero T
	var lastErr error
	delay := r.config.RetryDelay

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry (exponential backoff)
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(delay):
				delay = time.Duration(float64(delay) * r.config.RetryBackoff)
			}
		}

		// Try to execute the operation
		result, err := r.retriableOperation.Retry(ctx, attempt)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// If this is the last attempt, don't retry
		if attempt == r.config.MaxRetries {
			break
		}

		// Handle operation failure and get corrected parameters

		if err := r.retriableOperation.OnFailure(ctx, attempt, err); err != nil {
			return zero, err
		}
	}

	return zero, fmt.Errorf("operation failed after %d retries: %w", r.config.MaxRetries, lastErr)
}
