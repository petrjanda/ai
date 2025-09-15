package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPTool struct {
	Name_        string          `json:"name"`
	Description_ string          `json:"description"`
	InputSchema  json.RawMessage `json:"inputSchema"`
	OutputSchema json.RawMessage `json:"outputSchema"`

	session *mcp.ClientSession
}

func NewFromMCPTool(session *mcp.ClientSession, tool *mcp.Tool) *MCPTool {

	schemaBytes, err := json.Marshal(tool.InputSchema)
	if err != nil {
		panic(err)
	}

	outputSchemaBytes, err := json.Marshal(tool.OutputSchema)
	if err != nil {
		outputSchemaBytes = nil
	}

	return &MCPTool{
		Name_:        tool.Name,
		Description_: tool.Description,
		InputSchema:  schemaBytes,
		OutputSchema: outputSchemaBytes,
		session:      session,
	}
}

func (t *MCPTool) Name() string {
	return t.Name_
}

func (t *MCPTool) Description() string {
	return t.Description_
}

func (t *MCPTool) InputSchemaRaw() json.RawMessage {
	return t.InputSchema
}

func (t *MCPTool) OutputSchemaRaw() json.RawMessage {
	return t.OutputSchema
}

func (t *MCPTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	res, err := t.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      t.Name(),
		Arguments: args,
	})

	if err != nil {
		return nil, err
	}

	if len(res.Content) == 0 {
		return nil, fmt.Errorf("no content returned from tool")
	}

	last := res.Content[len(res.Content)-1]

	typed, ok := last.(*mcp.TextContent)
	if !ok {
		return nil, fmt.Errorf("unsupported MCP response content type %T, must be text", last)
	}

	return json.RawMessage(typed.Text), nil
}

func GetMCPTools(ctx context.Context, session *mcp.ClientSession) Toolbox {
	var tools_ Toolbox
	for tool := range session.Tools(ctx, nil) {
		tools_ = append(tools_, NewFromMCPTool(session, tool))
	}
	return tools_
}
