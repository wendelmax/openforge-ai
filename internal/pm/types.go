// Package pm (Provider Manager) handles auto-discovery, lifecycle, and routing
// for local LLM inference runtimes (OpenVINO, Ollama, llama.cpp, vLLM, LM Studio).
package pm

import (
	"context"
	"time"
)

// ProviderType identifies a local inference runtime.
type ProviderType string

const (
	ProviderOpenVINO  ProviderType = "openvino"
	ProviderOllama    ProviderType = "ollama"
	ProviderLlamaCpp  ProviderType = "llamacpp"
	ProviderVLLM      ProviderType = "vllm"
	ProviderLMStudio  ProviderType = "lmstudio"
)

// WorkloadType identifies the kind of inference task.
type WorkloadType string

const (
	WorkloadChat     WorkloadType = "chat"
	WorkloadCode     WorkloadType = "code"
	WorkloadEmbed    WorkloadType = "embed"
	WorkloadRerank   WorkloadType = "rerank"
	WorkloadPlan     WorkloadType = "plan"
)

// ProviderStatus describes the current state of a provider.
type ProviderStatus string

const (
	StatusUnknown     ProviderStatus = "unknown"
	StatusAvailable   ProviderStatus = "available"
	StatusUnavailable ProviderStatus = "unavailable"
	StatusStarting    ProviderStatus = "starting"
	StatusStopped     ProviderStatus = "stopped"
	StatusError       ProviderStatus = "error"
	StatusBusy        ProviderStatus = "busy"
)

// ProviderInfo holds static metadata about a provider.
type ProviderInfo struct {
	Type        ProviderType  `json:"type"`
	Name        string        `json:"name"`
	Version     string        `json:"version,omitempty"`
	Description string        `json:"description,omitempty"`
	Website     string        `json:"website,omitempty"`

	SupportedHardware  []string       `json:"supported_hardware,omitempty"`
	SupportedWorkloads []WorkloadType `json:"supported_workloads,omitempty"`
	Native             bool           `json:"native"`
	NeedsInstall       bool           `json:"needs_install"`
	NeedsGPU           bool           `json:"needs_gpu,omitempty"`
	DefaultPort        int            `json:"default_port,omitempty"`
	AutoStartable      bool           `json:"auto_startable,omitempty"`
}

// ProviderHealth contains health check results.
type ProviderHealth struct {
	Status  ProviderStatus `json:"status"`
	Latency time.Duration  `json:"latency,omitempty"`
	Models  int            `json:"models,omitempty"`
	Devices []string       `json:"devices,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// ToolDef describes a tool available to the model (OpenAI function format).
type ToolDef struct {
	Type     string      `json:"type"` // "function"
	Function FunctionDef `json:"function"`
}

// FunctionDef defines a function's name, description, and parameters schema.
type FunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ToolCall represents a model-initiated tool call.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

// FunctionCall contains the name and arguments for a tool invocation.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON-encoded args string
}

// ChatRequest is a provider-agnostic chat completion request.
type ChatRequest struct {
	Model       string         `json:"model"`
	Messages    []Message      `json:"messages"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	Temperature float32        `json:"temperature,omitempty"`
	TopP        float32        `json:"top_p,omitempty"`
	TopK        int            `json:"top_k,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
	Stop        []string       `json:"stop,omitempty"`
	Device      string         `json:"device,omitempty"`
	Tools       []ToolDef      `json:"tools,omitempty"`
	Options     map[string]any `json:"options,omitempty"`
}

// Message represents a single message in a chat conversation.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// ChatResponse is a provider-agnostic chat completion response.
type ChatResponse struct {
	Model     string       `json:"model"`
	Content   string       `json:"content"`
	ToolCalls []ToolCall   `json:"tool_calls,omitempty"`
	Usage     *Usage       `json:"usage,omitempty"`
	Provider  ProviderType `json:"provider"`
	Device    string       `json:"device,omitempty"`
}

// Token is a single streaming token from a ChatStream call.
type Token struct {
	Content    string     `json:"content"`
	Model      string     `json:"model,omitempty"`
	Done       bool       `json:"done"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	Usage      *Usage     `json:"usage,omitempty"`
}

// EmbedRequest is a provider-agnostic embedding request.
type EmbedRequest struct {
	Model  string   `json:"model"`
	Inputs []string `json:"inputs"`
	Device string   `json:"device,omitempty"`
}

// EmbedResponse is a provider-agnostic embedding response.
type EmbedResponse struct {
	Model      string       `json:"model"`
	Embeddings [][]float32  `json:"embeddings"`
	Usage      *Usage       `json:"usage,omitempty"`
	Provider   ProviderType `json:"provider"`
}

// Usage tracks token consumption.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Model describes a model available through a provider.
type Model struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Provider ProviderType `json:"provider"`
	Size     int64        `json:"size,omitempty"`
	Format   string       `json:"format,omitempty"`
	Loaded   bool         `json:"loaded,omitempty"`
}

// Provider defines the interface every local inference runtime must implement.
type Provider interface {
	Info() ProviderInfo
	Status(ctx context.Context) (*ProviderHealth, error)
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan Token, error)
	Embed(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error)
	ListModels(ctx context.Context) ([]Model, error)
	LoadModel(ctx context.Context, modelID string) error
	UnloadModel(ctx context.Context, modelID string) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// BuildToolDefs converts tool.Tool implementations to pm.ToolDef for provider requests.
func BuildToolDefs(tools []any) []ToolDef {
	defs := make([]ToolDef, 0, len(tools))
	for _, t := range tools {
		if td, ok := t.(interface {
			Name() string
			Description() string
			InputSchema() map[string]any
		}); ok {
			defs = append(defs, ToolDef{
				Type: "function",
				Function: FunctionDef{
					Name:        td.Name(),
					Description: td.Description(),
					Parameters:  td.InputSchema(),
				},
			})
		}
	}
	return defs
}
