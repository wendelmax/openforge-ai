package tool

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type TodoItem struct {
	Content    string `json:"content"`
	Status     string `json:"status"`      // pending, in_progress, completed
	ActiveForm string `json:"active_form,omitempty"`
}

type TodosTool struct {
	mu    sync.RWMutex
	todos []TodoItem
}

func NewTodosTool() *TodosTool { return &TodosTool{} }
func (t *TodosTool) Name() string { return "todos" }
func (t *TodosTool) Description() string { return "Manage a todo list. Args: {todos: [{content, status, active_form}]}" }
func (t *TodosTool) InputSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"todos": map[string]any{"type": "array", "items": map[string]any{
			"type": "object", "properties": map[string]any{
				"content": map[string]any{"type": "string"},
				"status":  map[string]any{"type": "string", "enum": []string{"pending", "in_progress", "completed"}},
				"active_form": map[string]any{"type": "string"},
			},
		}},
	}}
}

func icon(s string) string {
	switch s {
	case "completed": return "✅"
	case "in_progress": return "🔄"
	default: return "⬜"
	}
}

func (t *TodosTool) Run(_ context.Context, args map[string]any) (ToolResult, error) {
	t.todos = nil
	if raw, ok := args["todos"].([]any); ok {
		for _, item := range raw {
			if m, ok := item.(map[string]any); ok {
				c, _ := m["content"].(string)
				s, _ := m["status"].(string)
				af, _ := m["active_form"].(string)
				t.todos = append(t.todos, TodoItem{c, s, af})
			}
		}
	}

	var b strings.Builder
	b.WriteString("=== Todos ===\n\n")
	var p, ip, cp int
	for i, item := range t.todos {
		prefix := fmt.Sprintf("%d. %s ", i+1, icon(item.Status))
		if item.Status == "in_progress" && item.ActiveForm != "" {
			b.WriteString(prefix + item.ActiveForm + "\n")
		} else {
			b.WriteString(prefix + item.Content + "\n")
		}
		switch item.Status {
		case "pending": p++
		case "in_progress": ip++
		case "completed": cp++
		}
	}
	b.WriteString(fmt.Sprintf("\nPending: %d | In Progress: %d | Completed: %d", p, ip, cp))
	return ToolResult{Content: b.String()}, nil
}
