// Package mcp implements the Model Context Protocol for tool execution.
package mcp

// ProtocolVersion is the MCP protocol version supported.
const ProtocolVersion = "2025-03-26"

// JSON-RPC message types
type request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *ErrorObj   `json:"error,omitempty"`
}

// ErrorObj is a JSON-RPC error object.
type ErrorObj struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// InitializeParams sent by client during initialization.
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      Implementation     `json:"clientInfo"`
}

// InitializeResult returned by server after initialization.
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
}

// Implementation describes a protocol implementation.
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ClientCapabilities describes client capabilities.
type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Roots        *struct{}              `json:"roots,omitempty"`
	Sampling     *struct{}              `json:"sampling,omitempty"`
}

// ServerCapabilities describes server capabilities.
type ServerCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Tools        *ToolCapabilities      `json:"tools,omitempty"`
	Resources    *struct{}              `json:"resources,omitempty"`
	Prompts      *struct{}              `json:"prompts,omitempty"`
	Logging      *struct{}              `json:"logging,omitempty"`
}

// ToolCapabilities describes tool-related capabilities.
type ToolCapabilities struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Tool definition returned by tools/list.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema is a JSON Schema for tool parameters.
type InputSchema struct {
	Type       string                  `json:"type"`
	Properties map[string]PropertyDef  `json:"properties,omitempty"`
	Required   []string                `json:"required,omitempty"`
}

// PropertyDef defines a single property in a JSON Schema.
type PropertyDef struct {
	Type        string   `json:"type,omitempty"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// CallToolParams for tools/call.
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResult returned by tools/call.
type CallToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ContentItem is a piece of content in a tool result.
type ContentItem struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Resource *struct {
		URI  string `json:"uri"`
		Text string `json:"text"`
		MIMEType string `json:"mimeType,omitempty"`
	} `json:"resource,omitempty"`
}

// ListToolsResult returned by tools/list.
type ListToolsResult struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}
