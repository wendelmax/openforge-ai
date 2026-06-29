package tool

import (
	"context"

	"github.com/openforge-ai/openforge/internal/mcp"
)

type MCPAdapter struct {
	tool     mcp.MCPTool
	registry *mcp.Registry
	client   *mcp.Client
}

func NewMCPAdapter(tool mcp.MCPTool, registry *mcp.Registry, client *mcp.Client) *MCPAdapter {
	return &MCPAdapter{tool: tool, registry: registry, client: client}
}

func (a *MCPAdapter) Name() string {
	return "mcp_" + a.tool.ServerName + "_" + a.tool.Name
}

func (a *MCPAdapter) Description() string {
	return a.tool.Description
}

func (a *MCPAdapter) InputSchema() map[string]any {
	return a.tool.InputSchema
}

func (a *MCPAdapter) Run(ctx context.Context, args map[string]any) (ToolResult, error) {
	if a.client == nil {
		return ToolResult{Error: "MCP client not connected"}, nil
	}
	result, err := a.client.CallTool(ctx, a.tool.Name, args)
	if err != nil {
		return ToolResult{Error: err.Error()}, nil
	}
	var text string
	for _, item := range result.Content {
		if item.Type == "text" {
			text += item.Text
		}
	}
	return ToolResult{Content: text}, nil
}
