# Smart Device Selection

**Date:** 2026-06-26
**Status:** Draft
**Author:** OpenForge

## Problem

OpenForge currently selects a single device at startup using fixed priority (GPU > NPU > CPU). This is inflexible because:
- Users cannot choose a device per request
- Different workloads benefit from different devices (NPU for LLM, CPU for embeddings)
- No performance data drives device selection

## Design

### 1. Multi-Device Model Cache

Each model can be compiled for multiple devices simultaneously. The `loadedModel` struct stores a map of device → compiled model:

```go
type loadedModel struct {
    path             string
    ovModel          *C.ov_model_t
    compiledByDevice map[string]*C.ov_compiled_model_t
}
```

On first request to a device, compile and cache. Subsequent requests reuse the cached compiled model.

**Eviction:** No explicit eviction in v1. Each unique device+model pair adds ~1 compiled instance. With 3 devices and typical model counts, memory overhead is negligible.

### 2. Per-Request Device Field

Add `device` field to all API request types:

```go
type ChatRequest struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    Device      string    `json:"device,omitempty"`    // NEW
    MaxTokens   int       `json:"max_tokens,omitempty"`
    Temperature *float32  `json:"temperature,omitempty"`
    TopP        *float32  `json:"top_p,omitempty"`
    TopK        *int      `json:"top_k,omitempty"`
    Stream      bool      `json:"stream,omitempty"`
}

type CompletionRequest struct {
    Model       string   `json:"model"`
    Prompt      string   `json:"prompt"`
    Device      string   `json:"device,omitempty"`    // NEW
    MaxTokens   int      `json:"max_tokens,omitempty"`
    // ...
}

type EmbeddingRequest struct {
    Model string   `json:"model"`
    Input []string `json:"input"`
    Device string  `json:"device,omitempty"`          // NEW
}

type RerankRequest struct {
    Model     string   `json:"model"`
    Query     string   `json:"query"`
    Documents []string `json:"documents"`
    Device    string   `json:"device,omitempty"`      // NEW
}
```

Add `Device` to `InferenceParams`:

```go
type InferenceParams struct {
    Device      string  `json:"device,omitempty"`
    Temperature float32 `json:"temperature,omitempty"`
    TopK        int     `json:"top_k,omitempty"`
    TopP        float32 `json:"top_p,omitempty"`
    MaxTokens   int     `json:"max_tokens,omitempty"`
}
```

**Resolution order (first match wins):**
1. Request-level `device` field
2. Workload-based default (from config)
3. Auto-selected best device (GPU > NPU > CPU)

### 3. Workload-Based Defaults

New config section maps workload type to preferred device:

```yaml
models:
  workloads:
    chat: "NPU"
    completion: "NPU"
    embed: "CPU"
    rerank: "CPU"
```

Valid values: `"CPU"`, `"GPU"`, `"NPU"`, `"auto"` (uses GPU > NPU > CPU priority).

This is read at startup and can reference benchmark results if `benchmark.enabled=true`.

### 4. Auto-Benchmark

New config section:

```yaml
models:
  benchmark:
    enabled: true
    iterations: 3
    prompt: "The quick brown fox jumps over the lazy dog"
    max_tokens: 50
```

On startup, if `benchmark.enabled=true`, run a small synthetic workload on each device:
- **Text generation:** Generate `max_tokens` tokens, measure tokens/second
- **Embeddings:** Embed benchmark prompt, measure latency

Results stored in memory, exposed via `POST /v1/benchmark`:

```json
{
  "CPU": {
    "chat_tokens_per_sec": 45.2,
    "embed_latency_ms": 2.1
  },
  "NPU": {
    "chat_tokens_per_sec": 38.7,
    "embed_latency_ms": 5.3
  },
  "GPU": {
    "chat_tokens_per_sec": 62.1,
    "embed_latency_ms": 3.4
  }
}
```

If `workloads` is set to a device name (e.g., `chat: "NPU"`), it's fixed. If set to `"auto"`, the benchmark results determine the best device per workload.

### Data Flow

```
Request arrives
  device specified? ──yes──→ use that device
  no↓
  workloads table has mapping? ──yes──→ use mapped device
  no↓
  benchmark enabled? ──yes──→ use benchmarked best device
  no↓
  use static priority (GPU > NPU > CPU)
  ↓
  runtime.LoadModel(modelID, device)
    → check compiledByDevice[device]
    → cache hit? use cached
    → cache miss? compile, cache, use
  ↓
  Infer/Embed on selected compiled model
```

### Config Changes

Add to `internal/config/config.go`:

```go
type Config struct {
    // ... existing fields ...
    Models ModelsConfig `mapstructure:"models"`
}

type ModelsConfig struct {
    Path    string        `mapstructure:"path"`
    Default string        `mapstructure:"default"`
    Device  string        `mapstructure:"device"`
    Workloads WorkloadConfig `mapstructure:"workloads"`
    Benchmark BenchmarkConfig `mapstructure:"benchmark"`
}

type WorkloadConfig struct {
    Chat       string `mapstructure:"chat"`
    Completion string `mapstructure:"completion"`
    Embed      string `mapstructure:"embed"`
    Rerank     string `mapstructure:"rerank"`
}

type BenchmarkConfig struct {
    Enabled    bool   `mapstructure:"enabled"`
    Iterations int    `mapstructure:"iterations"`
    Prompt     string `mapstructure:"prompt"`
    MaxTokens  int    `mapstructure:"max_tokens"`
}
```

### Runtime Interface Changes

Pass device through the call chain:

```go
type Runtime interface {
    Infer(ctx context.Context, modelID string, prompt string, params InferenceParams) (*InferenceResult, error)
    InferStream(ctx context.Context, modelID string, prompt string, params InferenceParams) (<-chan string, error)
    Embed(ctx context.Context, modelID string, inputs []string, device string) (*EmbeddingResult, error)  // device param added
    // Rerank is server-side, handled by Embed + cosine similarity
    LoadModel(ctx context.Context, modelID, path, device string) error  // already has device
}
```

### Implementation Plan

1. **Phase 1: Multi-device cache** — Change `loadedModel` to support `compiledByDevice` map. Update `LoadModel` to compile per-device on demand.
2. **Phase 2: API device field** — Add `device` to request types, `InferenceParams`, plumb through server handlers and engine.
3. **Phase 3: Workload defaults** — Add config struct, wire into device resolution logic.
4. **Phase 4: Auto-benchmark** — Implement benchmark logic, config, and endpoint.
5. **Phase 5: Tests** — Unit tests for each phase.
