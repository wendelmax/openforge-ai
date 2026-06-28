// Package runtime defines core types and interfaces for the OpenForge AI runtime.
package runtime

import (
	"context"
)

// DeviceBenchmark holds benchmark results for a single device.
type DeviceBenchmark struct {
	ChatTokensPerSec float64 `json:"chat_tokens_per_sec"`
	EmbedLatencyMS   float64 `json:"embed_latency_ms"`
}

// BenchmarkResults maps device names to their benchmark results.
type BenchmarkResults map[string]DeviceBenchmark

// DeviceType represents the type of compute device (CPU, GPU, NPU).
type DeviceType string

// Device types supported by the runtime.
const (
	DeviceCPU DeviceType = "cpu"
	DeviceGPU DeviceType = "gpu"
	DeviceNPU DeviceType = "npu"
)

// Device represents a compute device with capabilities and availability.
type Device struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      DeviceType `json:"type"`
	Available bool       `json:"available"`
	Memory    int64      `json:"memory,omitempty"`
	Priority  int        `json:"priority"`
}

// ModelInfo describes a loaded or available model in the runtime.
type ModelInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Precision string `json:"precision"`
	Loaded    bool   `json:"loaded"`
	Size      int64  `json:"size"`
	Version   string `json:"version"`
}

// InferenceParams controls inference behavior (temperature, top-k, top-p, max tokens, device).
type InferenceParams struct {
	Temperature float32 `json:"temperature,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	TopP        float32 `json:"top_p,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Device      string  `json:"device,omitempty"`
}

// InferenceResult contains the output of a model inference.
type InferenceResult struct {
	Text     string  `json:"text"`
	Tokens   []int   `json:"tokens,omitempty"`
	LogProbs []float32 `json:"logprobs,omitempty"`
}

// EmbeddingResult contains embedding vectors for input texts.
type EmbeddingResult struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// Runtime defines the interface for AI model lifecycle management and inference.
type Runtime interface {
	ListDevices(ctx context.Context) ([]Device, error)
	LoadModel(ctx context.Context, modelID, path, device string) error
	UnloadModel(ctx context.Context, modelID string) error
	ListModels(ctx context.Context) ([]ModelInfo, error)
	Infer(ctx context.Context, modelID string, prompt string, params InferenceParams) (*InferenceResult, error)
	InferStream(ctx context.Context, modelID string, prompt string, params InferenceParams) (<-chan string, error)
	Embed(ctx context.Context, modelID string, inputs []string, device string) (*EmbeddingResult, error)
	Close(ctx context.Context) error

	// DeviceForWorkload resolves device: requestDevice (if non-empty) > workload default > global default > first available
	DeviceForWorkload(workload string, requestDevice string) string

	// Benchmark runs performance benchmarks for a loaded model across all compiled devices.
	Benchmark(ctx context.Context, modelID string, iterations int, prompt string, maxTokens int) (BenchmarkResults, error)
}
