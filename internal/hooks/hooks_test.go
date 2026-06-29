package hooks

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestEngine_PreToolUse_NoHooks(t *testing.T) {
	e := New(nil)
	results, allowed := e.RunPreToolUse(context.Background(), "bash", "{}")
	if !allowed {
		t.Error("expected allowed with no hooks")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestEngine_PreToolUse_Allow(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell hooks require sh")
	}
	e := New([]HookConfig{{
		Name: "test-allow", Run: "echo allow", Timeout: 5,
		Events: []HookEvent{PreToolUse},
	}})
	_, allowed := e.RunPreToolUse(context.Background(), "view", "{}")
	if !allowed {
		t.Error("expected allowed")
	}
}

func TestEngine_PreToolUse_Deny(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell hooks require sh")
	}
	e := New([]HookConfig{{
		Name: "test-deny", Run: "echo blocked >&2; exit 1", Timeout: 5,
		Events: []HookEvent{PreToolUse},
	}})
	results, allowed := e.RunPreToolUse(context.Background(), "bash", `{"command":"rm"}`)
	if allowed {
		t.Error("expected denied")
	}
	if len(results) != 1 || results[0].Allowed {
		t.Errorf("result should be denied: %+v", results)
	}
}

func TestEngine_PostToolUse(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell hooks require sh")
	}
	e := New([]HookConfig{{
		Name: "post-test", Run: "echo result: $CRUSH_TOOL_RESULT", Timeout: 5,
		Events: []HookEvent{PostToolUse},
	}})
	results := e.RunPostToolUse(context.Background(), "view", `{"path":"x"}`, "file content here")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestEngine_Parallel(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell hooks require sh")
	}
	e := New([]HookConfig{
		{Name: "h1", Run: "echo one", Timeout: 5, Events: []HookEvent{PreToolUse}},
		{Name: "h2", Run: "echo two", Timeout: 5, Events: []HookEvent{PreToolUse}},
		{Name: "h3", Run: "echo three", Timeout: 5, Events: []HookEvent{PreToolUse}},
	})
	results, allowed := e.RunPreToolUse(context.Background(), "bash", "{}")
	if !allowed {
		t.Error("expected allowed")
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestEngine_Timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell hooks require sh")
	}
	e := New([]HookConfig{{
		Name: "slow", Run: "sleep 10", Timeout: 2,
		Events: []HookEvent{PreToolUse},
	}})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	results, _ := e.RunPreToolUse(ctx, "bash", "{}")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Allowed {
		t.Error("timed-out hook should result in denial")
	}
}
