package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	pm "github.com/openforge-ai/openforge/internal/pm"
	"github.com/openforge-ai/openforge/runtime"
)

// OllamaProvider implements pm.Provider for the Ollama runtime.
type OllamaProvider struct {
	baseURL string
	client  *http.Client
	info    pm.ProviderInfo
}

func NewOllamaProvider(baseURL string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{MaxIdleConns: 4, IdleConnTimeout: 30 * time.Second, DisableKeepAlives: false},
		},
		info: pm.ProviderInfo{
			Type: pm.ProviderOllama, Name: "Ollama",
			Description: "Local LLM runtime via Ollama. Supports hundreds of GGUF models.",
			Website:     "https://ollama.com",
			SupportedHardware:  []string{"CPU", "GPU (CUDA)", "GPU (Vulkan)", "GPU (Metal)"},
			SupportedWorkloads: []pm.WorkloadType{pm.WorkloadChat, pm.WorkloadCode, pm.WorkloadEmbed},
			NeedsInstall: true, DefaultPort: 11434, AutoStartable: true,
		},
	}
}

func (p *OllamaProvider) Info() pm.ProviderInfo { return p.info }

func (p *OllamaProvider) Status(ctx context.Context) (*pm.ProviderHealth, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/api/tags", nil)
	if err != nil {
		return pm.HealthError(err), nil
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return pm.HealthUnavailable(err.Error()), nil
	}
	defer resp.Body.Close()
	latency := time.Since(start)
	if resp.StatusCode != http.StatusOK {
		h := pm.HealthError(fmt.Errorf("HTTP %d", resp.StatusCode))
		h.Latency = latency
		return h, nil
	}
	var tags struct{ Models []struct{ Name string `json:"name"` } `json:"models"` }
	h := pm.HealthAvailable()
	h.Latency = latency
	if json.NewDecoder(resp.Body).Decode(&tags) == nil {
		h.Models = len(tags.Models)
		h.Devices = []string{"CPU"}
	}
	return h, nil
}

// --- Ollama API types ---

type ollamaToolDef struct {
	Type     string            `json:"type"`
	Function ollamaFunctionDef `json:"function"`
}
type ollamaFunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type ollamaToolCall struct {
	ID       string             `json:"id,omitempty"`
	Type     string             `json:"type,omitempty"`
	Function ollamaFunctionCall `json:"function"`
}
type ollamaFunctionCall struct {
	Name      string `json:"name"`
	Arguments any    `json:"arguments"` // Ollama returns string or object depending on version
}

type ollamaMsg struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

type ollamaChatReq struct {
	Model    string          `json:"model"`
	Messages []ollamaMsg     `json:"messages"`
	Stream   bool            `json:"stream"`
	Tools    []ollamaToolDef `json:"tools,omitempty"`
	Options  map[string]any  `json:"options,omitempty"`
}

type ollamaChatResp struct {
	Model     string    `json:"model"`
	Message   ollamaMsg `json:"message"`
	Done      bool      `json:"done"`
	EvalCount int       `json:"eval_count,omitempty"`
}

// --- Conversions ---

func toOllamaTools(defs []pm.ToolDef) []ollamaToolDef {
	if len(defs) == 0 { return nil }
	out := make([]ollamaToolDef, len(defs))
	for i, d := range defs {
		out[i] = ollamaToolDef{
			Type: d.Type,
			Function: ollamaFunctionDef{Name: d.Function.Name, Description: d.Function.Description, Parameters: d.Function.Parameters},
		}
	}
	return out
}

func convertToolCalls(tcs []pm.ToolCall) []ollamaToolCall {
	if len(tcs) == 0 { return nil }
	out := make([]ollamaToolCall, len(tcs))
	for i, tc := range tcs {
		args := tc.Function.Arguments
		// Ollama expects arguments as object, not string
		var argsObj any
		if err := json.Unmarshal([]byte(args), &argsObj); err == nil {
			argsObj = args
		}
		out[i] = ollamaToolCall{ID: tc.ID, Type: tc.Type, Function: ollamaFunctionCall{Name: tc.Function.Name, Arguments: argsObj}}
	}
	return out
}

func parseOllamaToolCalls(tcs []ollamaToolCall) []pm.ToolCall {
	if len(tcs) == 0 { return nil }
	out := make([]pm.ToolCall, len(tcs))
	for i, tc := range tcs {
		var argsStr string
		switch v := tc.Function.Arguments.(type) {
		case string:
			argsStr = v
		case map[string]any:
			data, _ := json.Marshal(v)
			argsStr = string(data)
		default:
			argsStr = fmt.Sprint(v)
		}
		out[i] = pm.ToolCall{ID: tc.ID, Type: tc.Type, Function: pm.FunctionCall{Name: tc.Function.Name, Arguments: argsStr}}
	}
	return out
}

// --- Chat ---

func (p *OllamaProvider) Chat(ctx context.Context, req *pm.ChatRequest) (*pm.ChatResponse, error) {
	msgs := make([]ollamaMsg, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = ollamaMsg{Role: m.Role, Content: m.Content, ToolCalls: convertToolCalls(m.ToolCalls)}
	}
	body := ollamaChatReq{Model: req.Model, Messages: msgs, Tools: toOllamaTools(req.Tools), Options: map[string]any{"temperature": req.Temperature, "num_predict": req.MaxTokens}}
	payload, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("ollama chat: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama chat: HTTP %d: %s", resp.StatusCode, string(b))
	}
	var oResp ollamaChatResp
	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil { return nil, fmt.Errorf("ollama chat decode: %w", err) }
	return &pm.ChatResponse{Model: oResp.Model, Content: oResp.Message.Content, ToolCalls: parseOllamaToolCalls(oResp.Message.ToolCalls), Usage: &runtime.Usage{CompletionTokens: oResp.EvalCount}, Provider: pm.ProviderOllama}, nil
}

// --- ChatStream ---

func (p *OllamaProvider) ChatStream(ctx context.Context, req *pm.ChatRequest) (<-chan pm.Token, error) {
	msgs := make([]ollamaMsg, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = ollamaMsg{Role: m.Role, Content: m.Content, ToolCalls: convertToolCalls(m.ToolCalls)}
	}
	body := ollamaChatReq{Model: req.Model, Messages: msgs, Stream: true, Tools: toOllamaTools(req.Tools), Options: map[string]any{"temperature": req.Temperature, "num_predict": req.MaxTokens}}
	payload, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("ollama stream: %w", err) }
	ch := make(chan pm.Token, 64)
	go func() {
		defer resp.Body.Close()
		defer close(ch)
		dec := json.NewDecoder(resp.Body)
		for {
			var oResp ollamaChatResp
			if err := dec.Decode(&oResp); err != nil {
				if err != io.EOF { ch <- pm.Token{Content: fmt.Sprintf("\n[ollama error: %v]", err), Done: true} }
				return
			}
			tok := pm.Token{Content: oResp.Message.Content, Model: oResp.Model, Done: oResp.Done}
			if oResp.Done {
				tok.Usage = &runtime.Usage{CompletionTokens: oResp.EvalCount}
				tok.ToolCalls = parseOllamaToolCalls(oResp.Message.ToolCalls)
			}
			ch <- tok
			if oResp.Done { return }
		}
	}()
	return ch, nil
}

// --- Embed, Models, Lifecycle ---

type ollamaEmbedReq struct{ Model string; Prompt []string `json:"input"` }
type ollamaEmbedResp struct{ Embeddings [][]float32 `json:"embeddings"` }

func (p *OllamaProvider) Embed(ctx context.Context, req *pm.EmbedRequest) (*pm.EmbedResponse, error) {
	body := ollamaEmbedReq{Model: req.Model, Prompt: req.Inputs}
	payload, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/embed", bytes.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("ollama embed: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed: HTTP %d: %s", resp.StatusCode, string(b))
	}
	var oResp ollamaEmbedResp
	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil { return nil, fmt.Errorf("ollama embed decode: %w", err) }
	return &pm.EmbedResponse{Model: req.Model, Embeddings: oResp.Embeddings, Provider: pm.ProviderOllama}, nil
}

func (p *OllamaProvider) ListModels(ctx context.Context) ([]pm.Model, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/api/tags", nil)
	resp, err := p.client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	var tags struct{ Models []struct{ Name string `json:"name"`; Size int64 `json:"size,omitempty"` } `json:"models"` }
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil { return nil, err }
	models := make([]pm.Model, len(tags.Models))
	for i, m := range tags.Models { models[i] = pm.Model{ID: m.Name, Name: m.Name, Provider: pm.ProviderOllama, Size: m.Size, Format: "GGUF"} }
	return models, nil
}
func (p *OllamaProvider) LoadModel(ctx context.Context, modelID string) error {
	body := map[string]string{"model": modelID, "stream": "false"}
	payload, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/pull", bytes.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(httpReq)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { b, _ := io.ReadAll(resp.Body); return fmt.Errorf("ollama pull: HTTP %d: %s", resp.StatusCode, string(b)) }
	return nil
}
func (p *OllamaProvider) UnloadModel(ctx context.Context, _ string) error { return nil }
func (p *OllamaProvider) Start(ctx context.Context) error                  { return nil }
func (p *OllamaProvider) Stop(ctx context.Context) error                   { return nil }
