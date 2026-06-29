package agent

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/openforge-ai/openforge/internal/pm"
	"github.com/openforge-ai/openforge/internal/tool"
)

func TestParseToolCalls(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		text := `<<TOOL_CALL>>{"tool":"bash","args":{"command":"ls"}}<<END_TOOL>>`
		calls := ParseToolCalls(text)
		if len(calls) != 1 {
			t.Fatalf("expected 1 call, got %d", len(calls))
		}
		if calls[0].Tool != "bash" {
			t.Errorf("expected bash, got %s", calls[0].Tool)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		text := `<<TOOL_CALL>>{"tool":"bash","args":{}}<<END_TOOL>>
<<TOOL_CALL>>{"tool":"view","args":{"path":"x"}}<<END_TOOL>>`
		calls := ParseToolCalls(text)
		if len(calls) != 2 {
			t.Fatalf("expected 2 calls, got %d", len(calls))
		}
	})

	t.Run("none", func(t *testing.T) {
		if calls := ParseToolCalls("hello world"); len(calls) != 0 {
			t.Errorf("expected 0 calls, got %d", len(calls))
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		if calls := ParseToolCalls("<<TOOL_CALL>>not json<<END_TOOL>>"); len(calls) != 0 {
			t.Errorf("expected 0 calls for bad json, got %d", len(calls))
		}
	})
}

func TestBuildMessages(t *testing.T) {
	history := []pm.Message{
		{Role: "user", Content: "old question"},
		{Role: "assistant", Content: "old answer"},
	}
	toolMsgs := []pm.Message{{Role: "tool", Content: "tool output"}}

	msgs := BuildMessages("system prompt", "new question", history, toolMsgs)

	if len(msgs) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "system" || msgs[0].Content != "system prompt" {
		t.Error("first should be system")
	}
	if msgs[1].Role != "user" || msgs[1].Content != "old question" {
		t.Error("second should be history[0]")
	}
	if msgs[3].Role != "user" || msgs[3].Content != "new question" {
		t.Error("fourth should be user message")
	}
	if msgs[4].Role != "tool" {
		t.Error("last should be tool message")
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	reg := tool.NewRegistry()
	reg.Register(&mockTool{name: "mock", desc: "a mock tool"})

	tmpl := "Tools: {{.ToolDescriptions}}\nCWD: {{.WorkingDir}}"
	result := BuildSystemPrompt(tmpl, reg)

	if !strings.Contains(result, "mock") {
		t.Errorf("expected mock tool description, got: %s", result)
	}
	wd, _ := os.Getwd()
	if !strings.Contains(result, wd) {
		t.Errorf("expected working dir, got: %s", result)
	}
}

func TestCoderSystemPrompt(t *testing.T) {
	if CoderSystemPrompt == "" {
		t.Error("prompt should not be empty")
	}
	if !strings.Contains(CoderSystemPrompt, "<<TOOL_CALL>>") {
		t.Error("prompt should contain TOOL_CALL marker")
	}
	if !strings.Contains(CoderSystemPrompt, "{{.ToolDescriptions}}") {
		t.Error("prompt should contain ToolDescriptions placeholder")
	}
}

type mockTool struct {
	name string
	desc string
}

func (m *mockTool) Name() string                    { return m.name }
func (m *mockTool) Description() string              { return m.desc }
func (m *mockTool) InputSchema() map[string]any      { return nil }
func (m *mockTool) Run(_ context.Context, _ map[string]any) (tool.ToolResult, error) {
	return tool.ToolResult{Content: "ok"}, nil
}
