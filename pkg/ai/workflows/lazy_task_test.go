package workflows

import (
	"context"
	"testing"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLLM is a mock implementation of ai.LLM for testing
type MockLLM struct {
	mock.Mock
}

func (m *MockLLM) Invoke(ctx context.Context, request *ai.LLMRequest) (*ai.LLMResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*ai.LLMResponse), args.Error(1)
}

func TestLazyTask_BasicFunctionality(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{}

	// Create a mock response
	expectedResponse := &ai.LLMResponse{
		Messages: []ai.Message{ai.NewAssistantMessage("test response")},
	}

	// Create a callback that returns the expected response
	callback := func(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
		return expectedResponse, nil
	}

	// Create the lazy task
	lazyTask := NewLazyTask("lazy-test", callback)

	// Test basic properties
	assert.Equal(t, "lazy-test", lazyTask.Name())

	// Test cloning
	cloned := lazyTask.Clone().(*LazyTask)
	assert.Equal(t, lazyTask.id, cloned.id)
	assert.NotNil(t, cloned.callback)

	// Test invocation
	response, err := lazyTask.Invoke(ctx, mockLLM, ai.History{})

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestLazyTask_WithName(t *testing.T) {
	// Create a callback
	callback := func(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
		return &ai.LLMResponse{
			Messages: []ai.Message{ai.NewAssistantMessage("test response")},
		}, nil
	}

	// Create the lazy task
	lazyTask := NewLazyTask("original-name", callback)

	// Change the name
	renamedTask := lazyTask.WithName("new-name")

	// Verify the name was changed
	assert.Equal(t, "new-name", renamedTask.Name())

	// Verify the original task name is unchanged
	assert.Equal(t, "original-name", lazyTask.Name())
}

func TestLazyTask_WithRequestOpts(t *testing.T) {
	// Create a callback
	callback := func(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
		return &ai.LLMResponse{
			Messages: []ai.Message{ai.NewAssistantMessage("test response")},
		}, nil
	}

	// Create the lazy task
	lazyTask := NewLazyTask("lazy-test", callback)

	// Add request options (should be ignored)
	lazyTaskWithOpts := lazyTask.WithRequestOpts(
		ai.WithTemperature(0.5),
		ai.WithMaxCompletionTokens(1000),
	)

	// Verify it's still the same task (just cloned)
	assert.Equal(t, lazyTask.id, lazyTaskWithOpts.(*LazyTask).id)
	// Note: We can't compare function pointers directly, but we can verify the task is cloned
	assert.NotNil(t, lazyTaskWithOpts.(*LazyTask).callback)
}

func TestLazyTask_CallbackError(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{}

	// Create a callback that returns an error
	callback := func(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
		return nil, assert.AnError
	}

	// Create the lazy task
	lazyTask := NewLazyTask("lazy-test", callback)

	// Test invocation should return the error
	response, err := lazyTask.Invoke(ctx, mockLLM, ai.History{})

	// Verify error is returned
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, assert.AnError, err)
}

func TestLazyTask_CallbackReceivesCorrectParameters(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{}

	var receivedLLM ai.LLM
	var receivedHistory ai.History
	var receivedCtx context.Context

	// Create a callback that captures the parameters
	callback := func(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
		receivedCtx = ctx
		receivedLLM = llm
		receivedHistory = history
		return &ai.LLMResponse{
			Messages: []ai.Message{ai.NewAssistantMessage("test response")},
		}, nil
	}

	// Create the lazy task
	lazyTask := NewLazyTask("lazy-test", callback)

	// Create test history
	testHistory := ai.History{
		ai.NewUserMessage("test message"),
	}

	// Test invocation
	response, err := lazyTask.Invoke(ctx, mockLLM, testHistory)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Verify the callback received the correct parameters
	assert.Equal(t, ctx, receivedCtx)
	assert.Equal(t, mockLLM, receivedLLM)
	assert.Equal(t, testHistory, receivedHistory)
}
