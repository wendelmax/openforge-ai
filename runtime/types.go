// Package runtime defines request and response types for the OpenForge AI API.
package runtime

import "time"

// Message represents a single message in a chat conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// InferenceRequest contains parameters for running model inference.
type InferenceRequest struct {
	ModelID    string    `json:"model"`
	Prompt     string    `json:"prompt,omitempty"`
	Messages   []Message `json:"messages,omitempty"`
	System     string    `json:"system,omitempty"`
	MaxTokens  int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	TopK       int       `json:"top_k,omitempty"`
	TopP       float32   `json:"top_p,omitempty"`
	Stream     bool      `json:"stream,omitempty"`
	Stop       []string  `json:"stop,omitempty"`
}

// ChatRequest is the request format for the chat completions API.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Device      string    `json:"device,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature *float32  `json:"temperature,omitempty"`
	TopP        *float32  `json:"top_p,omitempty"`
	TopK        *int      `json:"top_k,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
}

// CompletionRequest is the request format for text completions.
type CompletionRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Device      string   `json:"device,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature *float32 `json:"temperature,omitempty"`
	TopP        *float32 `json:"top_p,omitempty"`
	TopK        *int     `json:"top_k,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// EmbeddingRequest is the request format for generating embeddings.
type EmbeddingRequest struct {
	Model  string   `json:"model"`
	Input  []string `json:"input"`
	Device string   `json:"device,omitempty"`
}

// RerankRequest is the request format for reranking documents against a query.
type RerankRequest struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	Device    string   `json:"device,omitempty"`
	TopN      int      `json:"top_n,omitempty"`
}

// ModelLoadRequest contains parameters for loading a model into memory.
type ModelLoadRequest struct {
	ModelID string `json:"model_id"`
	Device  string `json:"device,omitempty"`
}

// ModelUnloadRequest contains parameters for unloading a model from memory.
type ModelUnloadRequest struct {
	ModelID string `json:"model_id"`
}

// BenchmarkRequest contains parameters for benchmarking a model.
type BenchmarkRequest struct {
	Model      string `json:"model"`
	Device     string `json:"device,omitempty"`
	Iterations int    `json:"iterations,omitempty"`
}

// Usage reports token consumption for a request.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Timing reports performance timing metrics for a request.
type Timing struct {
	TTFT            string  `json:"ttft,omitempty"`
	Total           string  `json:"total,omitempty"`
	TokensPerSecond float64 `json:"tokens_per_second,omitempty"`
	LoadDuration    string  `json:"load_duration,omitempty"`
}

// ChatResponse is the response format for the chat completions API.
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   *Usage    `json:"usage,omitempty"`
	Timing  *Timing   `json:"timing,omitempty"`
}

// Choice represents a single completion choice in a chat response.
type Choice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"`
	Delta        *Message `json:"delta,omitempty"`
	FinishReason *string  `json:"finish_reason,omitempty"`
}

// CompletionResponse is the response format for the text completions API.
type CompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []CompletionChoice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
	Timing  *Timing  `json:"timing,omitempty"`
}

// CompletionChoice represents a single text completion choice.
type CompletionChoice struct {
	Index        int     `json:"index"`
	Text         string  `json:"text"`
	FinishReason *string `json:"finish_reason,omitempty"`
}

// EmbeddingResponse is the response format for the embeddings API.
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  *Usage          `json:"usage,omitempty"`
}

// EmbeddingData represents a single embedding vector with metadata.
type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// RerankResponse is the response format for the reranking API.
type RerankResponse struct {
	Object  string        `json:"object"`
	Model   string        `json:"model"`
	Results []RerankResult `json:"results"`
}

// RerankResult represents a single reranked document result.
type RerankResult struct {
	Index    int     `json:"index"`
	Score    float32 `json:"score"`
	Document string  `json:"document"`
}

// ModelStatus represents the current state of a loaded model.
type ModelStatus struct {
	Object string     `json:"object"`
	ID     string     `json:"id"`
	Status string     `json:"status"`
	Device string     `json:"device,omitempty"`
	Timing *Timing    `json:"timing,omitempty"`
}

// DeviceListResponse is the response format for the list devices endpoint.
type DeviceListResponse struct {
	Object string   `json:"object"`
	Data   []Device `json:"data"`
}

// ModelListResponse is the response format for the list models endpoint.
type ModelListResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// HealthResponse contains the health status of the runtime.
type HealthResponse struct {
	Status       string `json:"status"`
	Version      string `json:"version"`
	Uptime       string `json:"uptime"`
	ModelsLoaded int    `json:"models_loaded"`
	ActiveDevice string `json:"active_device"`
}

// BenchmarkResponse contains performance benchmark results for a model.
type BenchmarkResponse struct {
	Model           string  `json:"model"`
	Device          string  `json:"device"`
	TokensPerSecond float64 `json:"tokens_per_second"`
	TTFTMs          float64 `json:"ttft_ms"`
	LatencyP50Ms    float64 `json:"latency_p50_ms"`
	LatencyP95Ms    float64 `json:"latency_p95_ms"`
	LatencyP99Ms    float64 `json:"latency_p99_ms"`
	MemoryMB        int64   `json:"memory_mb"`
}

// APIError represents an API error with a code and message.
type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorResponse is the standard error response wrapper returned by the API.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// NewErrorResponse creates an ErrorResponse with the given error code and message.
func NewErrorResponse(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
		},
	}
}

func newUsage(promptTokens, completionTokens int) *Usage {
	return &Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
}

var startTime = time.Now()

func uptime() string {
	return time.Since(startTime).Round(time.Second).String()
}
