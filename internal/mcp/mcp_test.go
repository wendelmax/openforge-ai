package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolStruct(t *testing.T) {
	tool := Tool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"name": {Type: "string", Description: "The name"},
			},
			Required: []string{"name"},
		},
	}
	assert.Equal(t, "test-tool", tool.Name)
	assert.Equal(t, "object", tool.InputSchema.Type)
	assert.Contains(t, tool.InputSchema.Required, "name")
}

func TestCallToolResult(t *testing.T) {
	result := CallToolResult{
		Content: []ContentItem{
			{Type: "text", Text: "hello"},
		},
		IsError: false,
	}
	assert.Len(t, result.Content, 1)
	assert.Equal(t, "text", result.Content[0].Type)
	assert.Equal(t, "hello", result.Content[0].Text)
	assert.False(t, result.IsError)
}

func TestListToolsResult(t *testing.T) {
	result := ListToolsResult{
		Tools: []Tool{
			{Name: "tool1", Description: "first tool"},
			{Name: "tool2", Description: "second tool"},
		},
	}
	assert.Len(t, result.Tools, 2)
	assert.Equal(t, "tool1", result.Tools[0].Name)
}

func TestInitializeParams(t *testing.T) {
	params := InitializeParams{
		ProtocolVersion: ProtocolVersion,
		ClientInfo: Implementation{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}
	assert.Equal(t, "2025-03-26", params.ProtocolVersion)
	assert.Equal(t, "test-client", params.ClientInfo.Name)
}

func TestInitializeResult(t *testing.T) {
	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		ServerInfo: Implementation{
			Name:    "test-server",
			Version: "0.1.0",
		},
		Capabilities: ServerCapabilities{
			Tools: &ToolCapabilities{ListChanged: true},
		},
	}
	assert.True(t, result.Capabilities.Tools.ListChanged)
	assert.Equal(t, "test-server", result.ServerInfo.Name)
}

func TestContentItemTypes(t *testing.T) {
	textItem := ContentItem{Type: "text", Text: "plain text"}
	assert.Equal(t, "text", textItem.Type)

	resourceItem := ContentItem{
		Type: "resource",
		Resource: &struct {
			URI      string `json:"uri"`
			Text     string `json:"text"`
			MIMEType string `json:"mimeType,omitempty"`
		}{
			URI:  "file:///test.txt",
			Text: "file content",
		},
	}
	assert.Equal(t, "resource", resourceItem.Type)
	assert.Equal(t, "file:///test.txt", resourceItem.Resource.URI)
}

func TestErrorObj(t *testing.T) {
	err := ErrorObj{
		Code:    -32601,
		Message: "Method not found",
	}
	assert.Equal(t, -32601, err.Code)
	assert.Equal(t, "Method not found", err.Message)
}

func TestServerConfig(t *testing.T) {
	cfg := ServerConfig{
		Command: "node",
		Args:    []string{"server.js", "--port", "3000"},
		Env:     []string{"NODE_ENV=production"},
	}
	assert.Equal(t, "node", cfg.Command)
	assert.Contains(t, cfg.Args, "--port")
}
