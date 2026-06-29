package tool

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type GlobTool struct{}

func NewGlobTool() *GlobTool { return &GlobTool{} }
func (t *GlobTool) Name() string { return "glob" }
func (t *GlobTool) Description() string { return "Find files by glob pattern. Args: {pattern, path}" }
func (t *GlobTool) InputSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"pattern": map[string]any{"type": "string"}, "path": map[string]any{"type": "string"}}, "required": []string{"pattern"},
	}
}
func (t *GlobTool) Run(ctx context.Context, args map[string]any) (ToolResult, error) {
	pat, _ := args["pattern"].(string)
	if pat == "" { return ToolResult{Error: "pattern required"}, nil }
	sp, _ := args["path"].(string)
	if sp == "" { sp = "." }
	type fe struct{ path string; mod time.Time }
	var matches []fe
	filepath.Walk(sp, func(p string, fi os.FileInfo, _ error) error {
		if fi != nil && strings.HasPrefix(fi.Name(), ".") && p != sp {
			if fi.IsDir() { return filepath.SkipDir }
			return nil
		}
		if fi == nil || fi.IsDir() { return nil }
		rel, _ := filepath.Rel(sp, p)
		m, _ := filepath.Match(pat, rel)
		if !m { m, _ = filepath.Match(pat, fi.Name()) }
		if m {
			matches = append(matches, fe{p, fi.ModTime()})
			if len(matches) >= 100 { return filepath.SkipAll }
		}
		return nil
	})
	sort.Slice(matches, func(i, j int) bool { return matches[i].mod.After(matches[j].mod) })
	if len(matches) == 0 { return ToolResult{Content: "No files found"}, nil }
	lines := make([]string, len(matches))
	for i, m := range matches { lines[i] = m.path }
	return ToolResult{Content: strings.Join(lines, "\n")}, nil
}
