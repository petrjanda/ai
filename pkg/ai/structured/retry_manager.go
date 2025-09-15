package structured

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsynq/ai/pkg/ai"
	"github.com/pkg/errors"
)

type RetryManager struct {
	config *RetryConfig
	llm    ai.LLM
}

// NewRetryManager creates a new retry manager with structured LLM support
func NewRetryManager(llm ai.LLM, config *RetryConfig) *RetryManager {
	return &RetryManager{
		config: config,
		llm:    llm,
	}
}

// ExecuteWithRetry executes an operation with retry logic
func ExecuteWithRetry[T any](rm *RetryManager, ctx context.Context, operation RetryableOperation[T]) (T, error) {
	var zero T
	var lastErr error
	delay := rm.config.RetryDelay

	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry (exponential backoff)
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(delay):
				delay = time.Duration(float64(delay) * rm.config.RetryBackoff)
			}
		}

		// Try to execute the operation
		result, err := operation.Execute(ctx, attempt)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// If this is the last attempt, don't retry
		if attempt == rm.config.MaxRetries {
			break
		}

		// Handle operation failure and get corrected parameters

		if err := handleOperationFailure(rm, ctx, operation, attempt, err); err != nil {
			return zero, err
		}
	}

	return zero, fmt.Errorf("operation failed after %d retries: %w", rm.config.MaxRetries, lastErr)
}

func handleOperationFailure[T any](rm *RetryManager, ctx context.Context, operation RetryableOperation[T], attempt int, err error) error {
	// Get corrected parameters from the LLM
	correctedArgs, err := getCorrectedParameters(rm, ctx, operation.GetCorrectionPrompt(ctx, err), operation.GetSchema())
	if err != nil {
		return errors.Wrap(err, "failed to get corrected parameters")
	}

	// Update the operation with corrected parameters
	if err := operation.UpdateWithCorrection(ctx, correctedArgs); err != nil {
		return errors.Wrap(err, "failed to update operation with correction")
	}

	slog.Info("LLM provided corrected parameters",
		"attempt", attempt+1,
		"corrected_params", prettyJSON(correctedArgs),
	)

	return nil
}

// getCorrectedParameters uses the structured LLM to get corrected parameters for a failed operation
func getCorrectedParameters(rm *RetryManager, ctx context.Context, prompt string, schema json.RawMessage) (json.RawMessage, error) {
	errorMessage := ai.NewUserMessage(prompt)

	retryRequest := ai.NewLLMRequest(
		ai.WithHistory(ai.NewHistory(errorMessage)),
		ai.WithModel(ai.ModelId(rm.config.Model)),
	)

	// Create a structured LLM around the provided schema to ensure proper JSON output
	structuredLLM := NewLLM(schema, rm.llm)
	retryResponse, retryErr := structuredLLM.Invoke(ctx, retryRequest)
	if retryErr != nil {
		return nil, fmt.Errorf("failed to get corrected parameters: %w", retryErr)
	}

	// Extract the corrected parameters from the structured LLM response
	// Structured LLM returns tool calls with the corrected parameters as arguments
	if len(retryResponse.ToolCalls()) > 0 {
		toolCall := retryResponse.ToolCalls()[0]
		return toolCall.Args, nil
	}

	return nil, fmt.Errorf("structured LLM did not provide corrected parameters")
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
