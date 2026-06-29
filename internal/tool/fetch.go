package tool

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type FetchTool struct{ networkEnabled bool }

func NewFetchTool(network bool) *FetchTool { return &FetchTool{network} }
func (t *FetchTool) Name() string { return "fetch" }
func (t *FetchTool) Description() string { return "Fetch a URL. Args: {url, format}. Max 100KB. Requires network enabled." }
func (t *FetchTool) InputSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"url": map[string]any{"type": "string"}, "format": map[string]any{"type": "string", "enum": []string{"text", "markdown"}},
	}, "required": []string{"url"}}
}
func (t *FetchTool) Run(ctx context.Context, args map[string]any) (ToolResult, error) {
	if !t.networkEnabled { return ToolResult{Error: "network disabled"}, nil }
	url, _ := args["url"].(string)
	if url == "" { return ToolResult{Error: "url required"}, nil }
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("User-Agent", "OpenForge/1.0")
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil { return ToolResult{Error: fmt.Sprintf("fetch: %v", err)}, nil }
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 100*1024))
	var b strings.Builder
	b.WriteString(fmt.Sprintf("URL: %s\nStatus: %d\nSize: %d\n\n", url, resp.StatusCode, len(data)))
	b.Write(data)
	return ToolResult{Content: b.String()}, nil
}
