package tools

// ToolUsage represents how tools should be used in an LLM request
type ToolUsage interface {
	Type() ToolUsageType
}

// ToolUsageType represents the type of tool usage strategy
type ToolUsageType string

const (
	// ToolUsageAuto allows the LLM to automatically choose when to use tools (default behavior)
	ToolUsageAuto ToolUsageType = "auto"

	// ToolUsageForced forces the LLM to use a specific tool
	ToolUsageForced ToolUsageType = "forced"
)

// AutoToolUsage allows automatic tool selection (default behavior)
type AutoToolUsage struct{}

func (a AutoToolUsage) Type() ToolUsageType {
	return ToolUsageAuto
}

// ForcedToolUsage forces the use of a specific tool
type ForcedToolUsage struct {
	ToolName string `json:"tool_name"`
}

func (f ForcedToolUsage) Type() ToolUsageType {
	return ToolUsageForced
}

// Helper functions for creating tool usage options
func AutoToolSelection() ToolUsage {
	return &AutoToolUsage{}
}

func ForceTool(toolName string) ToolUsage {
	return &ForcedToolUsage{ToolName: toolName}
}
