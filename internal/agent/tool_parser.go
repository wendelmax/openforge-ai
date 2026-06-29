package agent

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type ToolCall struct {
	Tool string         `json:"tool"`
	Args map[string]any `json:"args"`
}

var toolCallRe = regexp.MustCompile(`<<TOOL_CALL>>(.*?)<<END_TOOL>>`)
var inlineCallRe = regexp.MustCompile(`<<([a-z_]+)>>\s*(\{.*\})\s*<<END_TOOL>>`)
var curlyCallRe = regexp.MustCompile(`<<([a-z_]+)(\{[^}]*\})>>`)
var funcCallRe = regexp.MustCompile(`<<([a-z_]+)\s*\(\s*(.*?)\s*\)\s*>>`)

func ParseToolCalls(text string) []ToolCall {
	if calls := tryStandard(text); len(calls) > 0 { return calls }
	if calls := tryInline(text); len(calls) > 0 { return calls }
	if calls := tryCurlyCall(text); len(calls) > 0 { return calls }
	return tryFuncCall(text)
}

func tryStandard(text string) []ToolCall {
	matches := toolCallRe.FindAllStringSubmatch(text, -1)
	calls := make([]ToolCall, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 { continue }
		var tc ToolCall
		if json.Unmarshal([]byte(m[1]), &tc) == nil && tc.Tool != "" {
			calls = append(calls, tc)
		}
	}
	return calls
}

func tryInline(text string) []ToolCall {
	matches := inlineCallRe.FindAllStringSubmatch(text, -1)
	calls := make([]ToolCall, 0, len(matches))
	for _, m := range matches {
		if len(m) < 3 { continue }
		toolName := strings.TrimSpace(m[1])
		argsStr := strings.TrimSpace(m[2])
		var args map[string]any
		if json.Unmarshal([]byte(argsStr), &args) == nil {
			if inner, ok := args["args"].(map[string]any); ok { args = inner }
			calls = append(calls, ToolCall{Tool: toolName, Args: args})
		}
	}
	return calls
}

func tryCurlyCall(text string) []ToolCall {
	matches := curlyCallRe.FindAllStringSubmatch(text, -1)
	calls := make([]ToolCall, 0, len(matches))
	for _, m := range matches {
		if len(m) < 3 { continue }
		toolName := strings.TrimSpace(m[1])
		curlyContent := strings.TrimSpace(m[2])
		// Remove outer braces: {"arg1", "arg2"} → "arg1", "arg2"
		inner := curlyContent[1 : len(curlyContent)-1]
		args := splitArgs(inner)
		args = mapPositionalArgs(toolName, args)
		calls = append(calls, ToolCall{Tool: toolName, Args: args})
	}
	return calls
}

func tryFuncCall(text string) []ToolCall {
	matches := funcCallRe.FindAllStringSubmatch(text, -1)
	calls := make([]ToolCall, 0, len(matches))
	for _, m := range matches {
		if len(m) < 3 { continue }
		toolName := strings.TrimSpace(m[1])
		argsRaw := strings.TrimSpace(m[2])
		args := splitArgs(argsRaw)
		args = mapPositionalArgs(toolName, args)
		calls = append(calls, ToolCall{Tool: toolName, Args: args})
	}
	return calls
}

func splitArgs(raw string) map[string]any {
	args := make(map[string]any)
	re := regexp.MustCompile(`"([^"]*)"`)
	parts := re.FindAllStringSubmatch(raw, -1)
	if len(parts) > 0 {
		for i, p := range parts {
			if len(p) >= 2 { args[fmt.Sprintf("arg%d", i)] = p[1] }
		}
		return args
	}
	for i, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(strings.Trim(p, `"'`))
		if p != "" { args[fmt.Sprintf("arg%d", i)] = p }
	}
	return args
}

func mapPositionalArgs(tool string, args map[string]any) map[string]any {
	switch tool {
	case "view":
		if v, ok := args["arg0"]; ok { return map[string]any{"path": v} }
	case "write":
		path, _ := args["arg0"].(string)
		content, _ := args["arg1"].(string)
		return map[string]any{"path": path, "content": content}
	case "edit":
		path, _ := args["arg0"].(string)
		old, _ := args["arg1"].(string)
		new, _ := args["arg2"].(string)
		return map[string]any{"path": path, "old_string": old, "new_string": new}
	case "grep", "glob":
		if v, ok := args["arg0"]; ok { return map[string]any{"pattern": v} }
	case "ls":
		if v, ok := args["arg0"]; ok { return map[string]any{"path": v} }
	case "bash":
		if v, ok := args["arg0"]; ok { return map[string]any{"command": v} }
	case "fetch":
		if v, ok := args["arg0"]; ok { return map[string]any{"url": v} }
	}
	return args
}
