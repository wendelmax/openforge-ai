package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type LsTool struct{}

func NewLsTool() *LsTool { return &LsTool{} }
func (t *LsTool) Name() string { return "ls" }
func (t *LsTool) Description() string { return "List directory as tree. Args: {path}. Max depth 3." }
func (t *LsTool) InputSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{"path": map[string]any{"type": "string"}}}
}
func (t *LsTool) Run(ctx context.Context, args map[string]any) (ToolResult, error) {
	dp, _ := args["path"].(string)
	if dp == "" {
		dp = "."
	}
	var b strings.Builder
	printTree(&b, dp, "", 0)
	return ToolResult{Content: b.String()}, nil
}
func printTree(b *strings.Builder, dir, prefix string, depth int) {
	if depth >= 3 {
		return
	}
	entries, _ := os.ReadDir(dir)
	var vis []os.DirEntry
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			vis = append(vis, e)
		}
	}
	sort.Slice(vis, func(i, j int) bool {
		if vis[i].IsDir() != vis[j].IsDir() {
			return vis[i].IsDir()
		}
		return vis[i].Name() < vis[j].Name()
	})
	for i, e := range vis {
		last := i == len(vis)-1
		conn := "├── "
		cp := prefix + "│   "
		if last {
			conn = "└── "
			cp = prefix + "    "
		}
		if e.IsDir() {
			fmt.Fprintf(b, "%s%s%s/\n", prefix, conn, e.Name())
			printTree(b, filepath.Join(dir, e.Name()), cp, depth+1)
		} else {
			fmt.Fprintf(b, "%s%s%s\n", prefix, conn, e.Name())
		}
	}
}
