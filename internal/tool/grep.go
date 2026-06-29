package tool

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type GrepTool struct{}

func NewGrepTool() *GrepTool { return &GrepTool{} }
func (t *GrepTool) Name() string { return "grep" }
func (t *GrepTool) Description() string { return "Search file contents by pattern. Args: {pattern, path, include}" }
func (t *GrepTool) InputSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"pattern": map[string]any{"type": "string"}, "path": map[string]any{"type": "string"}, "include": map[string]any{"type": "string"}},
		"required": []string{"pattern"},
	}
}
func (t *GrepTool) Run(ctx context.Context, args map[string]any) (ToolResult, error) {
	p, _ := args["pattern"].(string)
	if p == "" { return ToolResult{Error: "pattern required"}, nil }
	sp, _ := args["path"].(string)
	if sp == "" { sp = "." }
	inc, _ := args["include"].(string)
	if runtime.GOOS == "windows" { return t.grepWalk(ctx, p, sp, inc) }
	return t.grepExec(ctx, p, sp, inc)
}
func (t *GrepTool) grepExec(ctx context.Context, pattern, path, include string) (ToolResult, error) {
	ga := []string{"-rnI"}
	if include != "" { ga = append(ga, "--include="+include) }
	ga = append(ga, pattern, path)
	out, err := exec.CommandContext(ctx, "grep", ga...).CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 { return ToolResult{Content: "No matches"}, nil }
		return ToolResult{Content: string(out), Error: err.Error()}, nil
	}
	return ToolResult{Content: string(out)}, nil
}
func (t *GrepTool) grepWalk(ctx context.Context, pattern, path, include string) (ToolResult, error) {
	var results []string
	filepath.Walk(path, func(p string, fi os.FileInfo, _ error) error {
		if fi != nil && fi.IsDir() && strings.HasPrefix(fi.Name(), ".") && p != path { return filepath.SkipDir }
		if fi == nil || fi.IsDir() { return nil }
		if include != "" { m, _ := filepath.Match(include, fi.Name()); if !m { return nil } }
		data, err := os.ReadFile(p)
		if err != nil { return nil }
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, pattern) {
				results = append(results, fmt.Sprintf("%s:%d: %s", p, i+1, line))
				if len(results) >= 100 { return filepath.SkipAll }
			}
		}
		return nil
	})
	if len(results) == 0 { return ToolResult{Content: "No matches"}, nil }
	return ToolResult{Content: strings.Join(results, "\n")}, nil
}
