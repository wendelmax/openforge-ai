package hooks

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type HookEvent string

const (
	PreToolUse  HookEvent = "PreToolUse"
	PostToolUse HookEvent = "PostToolUse"
)

type HookConfig struct {
	Name    string      `json:"name"`
	Run     string      `json:"run"`
	Timeout int         `json:"timeout"`
	Events  []HookEvent `json:"events"`
}

type HookResult struct {
	Name     string `json:"name"`
	Allowed  bool   `json:"allowed"`
	Message  string `json:"message,omitempty"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output,omitempty"`
}

type Engine struct {
	hooks  []HookConfig
	logger *slog.Logger
}

func New(hooks []HookConfig) *Engine {
	return &Engine{hooks: hooks, logger: slog.Default().With("component", "hooks")}
}

func hasEvent(cfg HookConfig, event HookEvent) bool {
	for _, e := range cfg.Events {
		if e == event {
			return true
		}
	}
	return false
}

func (e *Engine) RunPreToolUse(ctx context.Context, toolName string, args string) ([]HookResult, bool) {
	results := e.runHooks(ctx, PreToolUse, toolName, args, "")
	allowed := true
	for _, r := range results {
		if !r.Allowed {
			allowed = false
		}
	}
	return results, allowed
}

func (e *Engine) RunPostToolUse(ctx context.Context, toolName string, args string, result string) []HookResult {
	return e.runHooks(ctx, PostToolUse, toolName, args, result)
}

func (e *Engine) runHooks(ctx context.Context, event HookEvent, toolName, args, toolResult string) []HookResult {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []HookResult

	for _, h := range e.hooks {
		if !hasEvent(h, event) {
			continue
		}
		wg.Add(1)
		go func(hook HookConfig) {
			defer wg.Done()
			res := e.executeHook(ctx, hook, toolName, args, toolResult)
			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}(h)
	}
	wg.Wait()
	return results
}

func (e *Engine) executeHook(ctx context.Context, hook HookConfig, toolName, args, toolResult string) HookResult {
	timeout := hook.Timeout
	if timeout <= 0 {
		timeout = 30
	}
	hookCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(hookCtx, "sh", "-c", hook.Run)
	cmd.Env = append(cmd.Environ(), "CRUSH_TOOL_NAME="+toolName, "CRUSH_TOOL_ARGS="+args)
	if toolResult != "" {
		cmd.Env = append(cmd.Env, "CRUSH_TOOL_RESULT="+toolResult)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := HookResult{Name: hook.Name}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = -1
			res.Message = err.Error()
			res.Allowed = false
			return res
		}
	}

	stdoutStr := strings.TrimSpace(stdout.String())
	stderrStr := strings.TrimSpace(stderr.String())

	switch {
	case res.ExitCode != 0:
		res.Allowed = false
		res.Message = stderrStr
		res.Output = stdoutStr
	case stderrStr != "":
		res.Allowed = false
		res.Message = stderrStr
		res.Output = stdoutStr
	case stdoutStr != "":
		res.Allowed = true
		res.Message = stdoutStr
	default:
		res.Allowed = true
	}
	return res
}

var _ = fmt.Sprintf
