package agent

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"testing"
// 	"time"
// )

// // mockLLM is a mock LLM that can simulate failures and retries
// type mockLLM struct {
// 	invokeCount int
// 	shouldFail  bool
// 	correctArgs json.RawMessage
// }

// func (m *mockLLM) Invoke(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
// 	m.invokeCount++

// 	if m.shouldFail && m.invokeCount == 1 {
// 		// First call fails, return error
// 		return nil, errors.New("mock LLM failure")
// 	}

// 	// Return corrected parameters in an assistant message (simplified formatter behavior)
// 	return &LLMResponse{
// 		Messages: []Message{
// 			&AssistantMessage{
// 				Content: string(m.correctArgs),
// 			},
// 		},
// 	}, nil
// }

// // mockTool is a mock tool that fails on specific arguments
// type mockTool struct {
// 	name        string
// 	shouldFail  bool
// 	correctArgs json.RawMessage
// }

// func (m *mockTool) Name() string        { return m.name }
// func (m *mockTool) Description() string { return "Mock tool for testing" }
// func (m *mockTool) InputSchemaRaw() json.RawMessage {
// 	return json.RawMessage(`{"type": "object"}`)
// }
// func (m *mockTool) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
// 	if m.shouldFail && string(args) != string(m.correctArgs) {
// 		return nil, errors.New("invalid arguments")
// 	}
// 	return json.RawMessage(`{"result": "success"}`), nil
// }

// func TestAgentRetryMechanism(t *testing.T) {
// 	ctx := context.Background()

// 	// Test case: Tool fails initially, then succeeds with corrected args
// 	t.Run("ToolFailureWithRetry", func(t *testing.T) {
// 		correctArgs := json.RawMessage(`{"param": "correct"}`)
// 		wrongArgs := json.RawMessage(`{"param": "wrong"}`)

// 		mockTool := &mockTool{
// 			name:        "test_tool",
// 			shouldFail:  true,
// 			correctArgs: correctArgs,
// 		}

// 		mockLLM := &mockLLM{
// 			correctArgs: correctArgs,
// 		}

// 		agent := NewAgent(mockLLM, WithMaxRetries(2))

// 		// Simulate a tool call response
// 		response := &LLMResponse{
// 			Messages: []Message{
// 				NewToolCallMessage(&ToolCall{
// 					Name: "test_tool",
// 					Args: wrongArgs,
// 				}),
// 			},
// 		}

// 		// Test the retry mechanism
// 		result, err := agent.(*Agent).CallTool(ctx, response.ToolCalls()[0])

// 		if err != nil {
// 			t.Errorf("Expected tool to succeed after retry, got error: %v", err)
// 		}

// 		if result == nil {
// 			t.Error("Expected non-nil result after successful retry")
// 		}

// 		// Verify the LLM was called to get corrected arguments
// 		if mockLLM.invokeCount < 1 {
// 			t.Error("Expected LLM to be called for retry")
// 		}
// 	})

// 	// Test case: Tool fails all retries
// 	t.Run("ToolFailureAllRetries", func(t *testing.T) {
// 		mockTool := &mockTool{
// 			name:       "test_tool",
// 			shouldFail: true,
// 			// No correct args, so it will always fail
// 		}

// 		// Mock LLM that doesn't provide corrected arguments
// 		mockLLM := &mockLLM{
// 			correctArgs: json.RawMessage(`{"param": "still_wrong"}`), // Still wrong args
// 		}

// 		agent := NewAgent(mockLLM, []Tool{mockTool}, WithMaxRetries(1))

// 		// Create a request with a tool call that will always fail
// 		request := NewLLMRequest([]Message{})
// 		request = request.Clone(WithTools(mockTool))

// 		response := &LLMResponse{
// 			Messages: []Message{
// 				NewToolCallMessage(&ToolCall{
// 					Name: "test_tool",
// 					Args: json.RawMessage(`{"param": "wrong"}`),
// 				}),
// 			},
// 		}

// 		// Test the retry mechanism
// 		result, err := agent.(*Agent).CallTool(ctx, response.ToolCalls()[0])

// 		if err == nil {
// 			t.Error("Expected tool to fail after all retries")
// 		}

// 		if result != nil {
// 			t.Error("Expected nil result after all retries failed")
// 		}

// 		// Verify error message contains retry information
// 		if err.Error() == "" {
// 			t.Error("Expected error message to contain retry information")
// 		}
// 	})

// 	// Test case: Custom retry configuration
// 	t.Run("CustomRetryConfiguration", func(t *testing.T) {
// 		mockTool := &mockTool{
// 			name: "test_tool",
// 		}

// 		mockLLM := &mockLLM{}

// 		agent := NewAgent(mockLLM, []Tool{mockTool},
// 			WithMaxRetries(5),
// 			WithRetryDelay(50*time.Millisecond),
// 			WithRetryBackoff(1.5),
// 		)

// 		agentStruct := agent.(*Agent)

// 		if agentStruct.maxRetries != 5 {
// 			t.Errorf("Expected maxRetries to be 5, got %d", agentStruct.maxRetries)
// 		}

// 		if agentStruct.retryDelay != 50*time.Millisecond {
// 			t.Errorf("Expected retryDelay to be 50ms, got %v", agentStruct.retryDelay)
// 		}

// 		if agentStruct.retryBackoff != 1.5 {
// 			t.Errorf("Expected retryBackoff to be 1.5, got %f", agentStruct.retryBackoff)
// 		}
// 	})

// 	// Test case: Context cancellation during retry
// 	t.Run("ContextCancellationDuringRetry", func(t *testing.T) {
// 		mockTool := &mockTool{
// 			name:       "test_tool",
// 			shouldFail: true,
// 		}

// 		mockLLM := &mockLLM{}

// 		agent := NewAgent(mockLLM, []Tool{mockTool},
// 			WithMaxRetries(10),
// 			WithRetryDelay(100*time.Millisecond),
// 		)

// 		// Create a context that gets cancelled after a short delay
// 		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
// 		defer cancel()

// 		response := &LLMResponse{
// 			Messages: []Message{
// 				NewToolCallMessage(&ToolCall{
// 					Name: "test_tool",
// 					Args: json.RawMessage(`{"param": "wrong"}`),
// 				}),
// 			},
// 		}

// 		// Test the retry mechanism with cancelled context
// 		result, err := agent.(*Agent).CallTool(ctx, response.ToolCalls()[0])

// 		if err == nil {
// 			t.Error("Expected context cancellation error")
// 		}

// 		if result != nil {
// 			t.Error("Expected nil result after context cancellation")
// 		}
// 	})
// }
