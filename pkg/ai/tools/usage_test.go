package tools

import (
	"testing"
)

func TestToolUsage(t *testing.T) {
	tests := []struct {
		name      string
		toolUsage ToolUsage
		expected  ToolUsageType
	}{
		{
			name:      "Auto tool selection",
			toolUsage: AutoToolSelection(),
			expected:  ToolUsageAuto,
		},
		{
			name:      "Forced tool",
			toolUsage: ForceTool("calculator"),
			expected:  ToolUsageForced,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.toolUsage.Type() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.toolUsage.Type())
			}
		})
	}
}

func TestForcedToolUsage(t *testing.T) {
	toolUsage := ForceTool("calculator")

	if forced, ok := toolUsage.(*ForcedToolUsage); !ok {
		t.Error("expected ForcedToolUsage type")
	} else if forced.ToolName != "calculator" {
		t.Errorf("expected tool name 'calculator', got '%s'", forced.ToolName)
	}
}
