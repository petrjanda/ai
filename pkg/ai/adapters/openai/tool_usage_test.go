package openai

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/getsynq/ai/pkg/ai/tools"
)

// mockTool is a simple mock implementation for testing
type mockTool struct {
	name string
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Description() string { return "Mock tool for testing" }
func (m *mockTool) InputSchemaRaw() json.RawMessage {
	return json.RawMessage(`{"type": "object"}`)
}
func (m *mockTool) OutputSchemaRaw() json.RawMessage {
	return json.RawMessage(`{"type": "object"}`)
}
func (m *mockTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	return json.RawMessage(`{"result": "mock"}`), nil
}

func TestConvertToolUsage(t *testing.T) {
	tests := []struct {
		name      string
		toolUsage tools.ToolUsage
		tools     []tools.Tool
		expectNil bool
	}{
		{
			name:      "Auto tool selection should return nil",
			toolUsage: tools.AutoToolSelection(),
			tools:     []tools.Tool{},
			expectNil: true,
		},
		{
			name:      "Forced tool should return tool choice",
			toolUsage: tools.ForceTool("calculator"),
			tools:     []tools.Tool{&mockTool{name: "calculator"}},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertToolUsage(tt.toolUsage, tt.tools)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectNil && result != nil {
				t.Errorf("expected nil result, got %v", result)
			}

			if !tt.expectNil && result == nil {
				t.Errorf("expected non-nil result, got nil")
			}
		})
	}
}
