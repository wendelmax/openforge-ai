package tool

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var dangerousPatterns = []string{
	"rm -rf /", "sudo rm", "chmod 777 /", "mkfs.", "dd if=",
	":(){ :|:& };:", "> /dev/sda", "wget -O- | sh", "curl | sh", "curl | bash",
}

type BashTool struct{}

func NewBashTool() *BashTool { return &BashTool{} }
func (t *BashTool) Name() string { return "bash" }
func (t *BashTool) Description() string {
	return "Execute a shell command. Rejects dangerous patterns. 30s timeout. Args: {command, workdir}"
}

func (t *BashTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{"type": "string"},
			"workdir": map[string]any{"type": "string"},
		},
		"required": []string{"command"},
	}
}

func (t *BashTool) Run(ctx context.Context, args map[string]any) (ToolResult, error) {
	cmd, _ := args["command"].(string)
	if cmd == "" {
		return ToolResult{Error: "command is required"}, nil
	}
	lower := strings.ToLower(cmd)
	for _, p := range dangerousPatterns {
		if strings.Contains(lower, p) {
			return ToolResult{Error: fmt.Sprintf("dangerous: %s", p)}, nil
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.CommandContext(ctx, "cmd", "/c", cmd)
	} else {
		c = exec.CommandContext(ctx, "sh", "-c", cmd)
	}
	if wd, ok := args["workdir"].(string); ok && wd != "" {
		c.Dir = wd
	}
	out, err := c.CombinedOutput()
	if err != nil {
		return ToolResult{Content: string(out), Error: err.Error()}, nil
	}
	return ToolResult{Content: string(out)}, nil
}
