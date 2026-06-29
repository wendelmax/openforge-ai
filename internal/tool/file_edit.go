package tool

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type EditTool struct{}

func NewEditTool() *EditTool { return &EditTool{} }
func (t *EditTool) Name() string { return "edit" }
func (t *EditTool) Description() string { return "Edit file by exact find-replace. Args: {path, old_string, new_string}" }
func (t *EditTool) InputSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"path": map[string]any{"type": "string"}, "old_string": map[string]any{"type": "string"}, "new_string": map[string]any{"type": "string"}},
		"required": []string{"path", "old_string", "new_string"},
	}
}
func (t *EditTool) Run(ctx context.Context, args map[string]any) (ToolResult, error) {
	path, _ := args["path"].(string)
	oldS, _ := args["old_string"].(string)
	newS, _ := args["new_string"].(string)
	if path == "" { return ToolResult{Error: "path is required"}, nil }
	data, err := os.ReadFile(path)
	if err != nil { return ToolResult{Error: fmt.Sprintf("read: %v", err)}, nil }
	content := string(data)
	idx := strings.Index(content, oldS)
	if idx < 0 { return ToolResult{Error: "old_string not found"}, nil }
	nc := content[:idx] + newS + content[idx+len(oldS):]
	os.WriteFile(path, []byte(nc), 0644)
	return ToolResult{Content: fmt.Sprintf("Edited %s — replaced occurrence at offset %d", path, idx)}, nil
}
