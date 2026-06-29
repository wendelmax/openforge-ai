package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openforge-ai/openforge/internal/pm"
	"github.com/openforge-ai/openforge/internal/tool"
)

const maxToolIterations = 10

type Agent struct {
	cfg      AgentConfig
	registry *tool.Registry
	history  []pm.Message
	busy     bool
}

type AgentConfig struct {
	Model        string
	MaxTokens    int
	Temperature  float32
	Provider     pm.Provider
	Tools        []tool.Tool
	SystemPrompt string
}

func New(cfg AgentConfig) *Agent {
	reg := tool.NewRegistry()
	for _, t := range cfg.Tools {
		reg.Register(t)
	}
	return &Agent{cfg: cfg, registry: reg, history: make([]pm.Message, 0)}
}

func (a *Agent) Run(ctx context.Context, userMessage string, streamFn func(string)) (string, error) {
	a.busy = true
	defer func() { a.busy = false }()

	sp := BuildSystemPrompt(a.cfg.SystemPrompt, a.registry)
	a.history = append(a.history, pm.Message{Role: "user", Content: userMessage})

	var toolMsgs []pm.Message
	for i := 0; i < maxToolIterations; i++ {
		msgs := BuildMessages(sp, "", a.history, toolMsgs)
		toolMsgs = nil

		req := &pm.ChatRequest{Model: a.cfg.Model, Messages: msgs, MaxTokens: a.cfg.MaxTokens, Temperature: a.cfg.Temperature, Tools: a.buildToolDefs()}
		resp, err := a.cfg.Provider.Chat(ctx, req)
		if err != nil { return "", fmt.Errorf("chat: %w", err) }

		if len(resp.ToolCalls) > 0 {
			a.history = append(a.history, assistantMsg(resp.Content, resp.ToolCalls))
			for _, tc := range resp.ToolCalls { toolMsgs = append(toolMsgs, a.executeToolCall(ctx, tc)) }
			continue
		}

		calls := ParseToolCalls(resp.Content)
		if len(calls) > 0 {
			a.history = append(a.history, pm.Message{Role: "assistant", Content: resp.Content})
			for _, call := range calls {
				t, ok := a.registry.Get(call.Tool)
				if !ok { toolMsgs = append(toolMsgs, pm.Message{Role: "tool", Content: fmt.Sprintf("unknown tool: %s", call.Tool)}); continue }
				result, _ := t.Run(ctx, call.Args)
				toolMsgs = append(toolMsgs, pm.Message{Role: "tool", Content: formatToolOutput(call.Tool, result.Content, result.Error)})
			}
			continue
		}

		a.history = append(a.history, pm.Message{Role: "assistant", Content: resp.Content})
		if streamFn != nil { streamFn(resp.Content) }
		return resp.Content, nil
	}
	return "", fmt.Errorf("exceeded max tool iterations (%d)", maxToolIterations)
}

func (a *Agent) RunStream(ctx context.Context, userMessage string, tokenFn func(string), toolFn func(name, args string)) (string, error) {
	a.busy = true
	defer func() { a.busy = false }()

	sp := BuildSystemPrompt(a.cfg.SystemPrompt, a.registry)
	a.history = append(a.history, pm.Message{Role: "user", Content: userMessage})

	var toolMsgs []pm.Message
	toolDefs := a.buildToolDefs()

	for i := 0; i < maxToolIterations; i++ {
		msgs := BuildMessages(sp, "", a.history, toolMsgs)
		toolMsgs = nil

		req := &pm.ChatRequest{Model: a.cfg.Model, Messages: msgs, MaxTokens: a.cfg.MaxTokens, Temperature: a.cfg.Temperature, Stream: true, Tools: toolDefs}
		ch, err := a.cfg.Provider.ChatStream(ctx, req)
		if err != nil { return "", fmt.Errorf("stream: %w", err) }

		var content strings.Builder
		var toolCalls []pm.ToolCall
		for tok := range ch {
			if tok.Done { toolCalls = tok.ToolCalls; break }
			content.WriteString(tok.Content)
			if tokenFn != nil { tokenFn(tok.Content) }
		}

		if len(toolCalls) > 0 {
			a.history = append(a.history, assistantMsg(content.String(), toolCalls))
			for _, tc := range toolCalls { toolMsgs = append(toolMsgs, a.executeToolCall(ctx, tc)) }
			continue
		}

		ct := content.String()
		calls := ParseToolCalls(ct)
		if len(calls) > 0 {
			a.history = append(a.history, pm.Message{Role: "assistant", Content: ct})
			for _, call := range calls {
				if toolFn != nil {
					argsJSON, _ := json.Marshal(call.Args); toolFn(call.Tool, string(argsJSON))
				}
				t, ok := a.registry.Get(call.Tool)
				if !ok { toolMsgs = append(toolMsgs, pm.Message{Role: "tool", Content: fmt.Sprintf("unknown: %s", call.Tool)}); continue }
				result, _ := t.Run(ctx, call.Args)
				toolMsgs = append(toolMsgs, pm.Message{Role: "tool", Content: formatToolOutput(call.Tool, result.Content, result.Error)})
			}
			continue
		}

		a.history = append(a.history, pm.Message{Role: "assistant", Content: ct})
		return ct, nil
	}
	return "", fmt.Errorf("exceeded max tool iterations (%d)", maxToolIterations)
}

func (a *Agent) executeToolCall(ctx context.Context, tc pm.ToolCall) pm.Message {
	t, ok := a.registry.Get(tc.Function.Name)
	if !ok { return pm.Message{Role: "tool", ToolCallID: tc.ID, Content: fmt.Sprintf("unknown tool: %s", tc.Function.Name)} }
	var args map[string]any
	json.Unmarshal([]byte(tc.Function.Arguments), &args)
	result, _ := t.Run(ctx, args)
	return pm.Message{Role: "tool", ToolCallID: tc.ID, Content: formatToolOutput(tc.Function.Name, result.Content, result.Error)}
}

func (a *Agent) buildToolDefs() []pm.ToolDef {
	defs := make([]pm.ToolDef, 0)
	for _, t := range a.cfg.Tools { defs = append(defs, pm.ToolDef{Type: "function", Function: pm.FunctionDef{Name: t.Name(), Description: t.Description(), Parameters: t.InputSchema()}}) }
	return defs
}

func (a *Agent) Reset() { a.history = make([]pm.Message, 0) }

func assistantMsg(content string, toolCalls []pm.ToolCall) pm.Message {
	return pm.Message{Role: "assistant", Content: content, ToolCalls: toolCalls}
}

// SaveSession persists conversation history as JSON.
func (a *Agent) SaveSession(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil { return fmt.Errorf("save session: %w", err) }
	data, err := json.MarshalIndent(a.history, "", "  ")
	if err != nil { return fmt.Errorf("marshal history: %w", err) }
	if err := os.WriteFile(path, data, 0644); err != nil { return fmt.Errorf("write session: %w", err) }
	return nil
}

// LoadSession loads conversation history from a JSON file.
func (a *Agent) LoadSession(path string) error {
	data, err := os.ReadFile(path)
	if err != nil { return fmt.Errorf("read session: %w", err) }
	var msgs []pm.Message
	if err := json.Unmarshal(data, &msgs); err != nil { return fmt.Errorf("unmarshal history: %w", err) }
	a.history = msgs
	return nil
}
