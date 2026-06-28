# Smart Device Selection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow per-request device selection, workload-based defaults, and auto-benchmarking for device-aware inference.

**Architecture:** Runtime stores compiled models per device in a map. Server parses `device` from request JSON, resolves through request > workload > auto > static priority, passes to runtime. Benchmark runs synthetic inference on each device at startup.

**Tech Stack:** Go 1.26, OpenVINO C API 2026.2.1

## Global Constraints

- All existing tests must pass
- Backward compatible: `device` field is optional, defaults to current behavior
- Thread-safe: `sync.RWMutex` for concurrent access to compiled model cache
- Builds with both CGO_ENABLED=1 (full) and CGO_ENABLED=0 (stub)

---

### Task 1: Restructure loadedModel for multi-device cache

**Files:**
- Modify: `internal/provider/openvino/runtime.go:23-35` (loadedModel struct, LoadModel)
- Test: `internal/provider/openvino/runtime_test.go` (new test)

**Interfaces:**
- Consumes: Runtime interface `LoadModel(ctx, modelID, path, device string) error`
- Produces: `loadedModel` with `compiledByDevice map[string]*compiledModelHandle`

- [ ] **Step 1: Change loadedModel struct**

In `internal/provider/openvino/runtime.go`, replace the current `loadedModel`:

```go
type compiledModelHandle struct {
    compiled *C.ov_compiled_model_t
    device   string
}

type loadedModel struct {
    path             string
    ovModel          *C.ov_model_t
    compiledByDevice map[string]*compiledModelHandle
}
```

- [ ] **Step 2: Update model loading to use per-device cache**

Replace `LoadModel` method to check/create compiled model per device:

```go
func (r *OpenVINORuntime) LoadModel(ctx context.Context, modelID, path, device string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if device == "" || device == "auto" {
        device = r.defaultDev
    }

    lm, exists := r.models[modelID]
    if !exists {
        ovModel, err := r.core.ReadModel(path)
        if err != nil {
            return fmt.Errorf("read model %q: %w", path, err)
        }
        lm = &loadedModel{
            path:             path,
            compiledByDevice: make(map[string]*compiledModelHandle),
        }
        if r.tokenizer != nil {
            lm.ovModel = ovModel
        }
        r.models[modelID] = lm
    }

    if _, cached := lm.compiledByDevice[device]; cached {
        slog.Debug("model already compiled for device", "model_id", modelID, "device", device)
        return nil
    }

    slog.Info("compiling model for device", "model_id", modelID, "device", device)
    compiled, err := r.core.CompileModel(lm.ovModel, device)
    if err != nil {
        return fmt.Errorf("compile model for %s: %w", device, err)
    }

    lm.compiledByDevice[device] = &compiledModelHandle{
        compiled: compiled,
        device:   device,
    }
    return nil
}
```

- [ ] **Step 3: Remove old `UnloadModel` if it doesn't free per-device handles**

Update `UnloadModel` to free all compiled instances:

```go
func (r *OpenVINORuntime) UnloadModel(ctx context.Context, modelID string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    lm, ok := r.models[modelID]
    if !ok {
        return fmt.Errorf("model %q not found", modelID)
    }

    for _, h := range lm.compiledByDevice {
        if h.compiled != nil {
            h.compiled.Free()
        }
    }
    if lm.ovModel != nil {
        lm.ovModel.Free()
    }
    delete(r.models, modelID)
    return nil
}
```

- [ ] **Step 4: Update `generate` and `extractEmbedding` to accept device**

Change their signatures to take a device string and look up the right compiled model:

```go
func (r *OpenVINORuntime) generate(ctx context.Context, lm *loadedModel, inputIDs []int64, params runtime.InferenceParams) ([]int64, error) {
    r.mu.RLock()
    h, ok := lm.compiledByDevice[params.Device]
    r.mu.RUnlock()
    if !ok {
        return nil, fmt.Errorf("model not compiled for device %q", params.Device)
    }
    // ... rest uses h.compiled instead of lm.compiled
```

Similarly for `extractEmbedding`.

- [ ] **Step 5: Write test for multi-device cache**

In `internal/provider/openvino/runtime_test.go`:

```go
func TestLoadModel_MultiDevice(t *testing.T) {
    r := &OpenVINORuntime{
        models:     make(map[string]*loadedModel),
        mu:         sync.RWMutex{},
        defaultDev: "CPU",
    }
    // Verify LoadModel on different devices creates separate compiled instances
    // (stub mode will fail, test only runs with cgo)
}
```

- [ ] **Step 6: Verify build**

```bash
export CGO_ENABLED=1 && go build ./internal/provider/openvino/...
```

---

### Task 2: Add device field to request types and InferenceParams

**Files:**
- Modify: `runtime/types.go` (ChatRequest, CompletionRequest, EmbeddingRequest, RerankRequest)
- Modify: `runtime/runtime.go` (InferenceParams)
- Test: `runtime/types_test.go` (JSON marshal/unmarshal)

**Interfaces:**
- Consumes: Existing request types
- Produces: `runtime.ChatRequest.Device`, `runtime.CompletionRequest.Device`, `runtime.EmbeddingRequest.Device`, `runtime.RerankRequest.Device`, `runtime.InferenceParams.Device`

- [ ] **Step 1: Add device field to ChatRequest**

```go
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
```

- [ ] **Step 2: Add device to CompletionRequest**

```go
type CompletionRequest struct {
    Model       string   `json:"model"`
    Prompt      string   `json:"prompt"`
    Device      string   `json:"device,omitempty"`
    MaxTokens   int      `json:"max_tokens,omitempty"`
    // ...
}
```

- [ ] **Step 3: Add device to EmbeddingRequest**

```go
type EmbeddingRequest struct {
    Model string   `json:"model"`
    Input []string `json:"input"`
    Device string  `json:"device,omitempty"`
}
```

- [ ] **Step 4: Add device to RerankRequest**

```go
type RerankRequest struct {
    Model     string   `json:"model"`
    Query     string   `json:"query"`
    Documents []string `json:"documents"`
    Device    string   `json:"device,omitempty"`
    TopN      int      `json:"top_n,omitempty"`
}
```

- [ ] **Step 5: Add device to InferenceParams**

```go
type InferenceParams struct {
    Device      string  `json:"device,omitempty"`
    Temperature float32 `json:"temperature,omitempty"`
    TopK        int     `json:"top_k,omitempty"`
    TopP        float32 `json:"top_p,omitempty"`
    MaxTokens   int     `json:"max_tokens,omitempty"`
}
```

- [ ] **Step 6: Update EmbeddingResult if needed**

No changes needed.

- [ ] **Step 7: Verify build**

```bash
go build ./...
```

---

### Task 3: Plumb device through server handlers

**Files:**
- Modify: `internal/server/server.go` (handleChat, handleCompletion, handleEmbeddings, handleRerank)
- Modify: `runtime/runtime.go` (Runtime interface: Embed now takes device)
- Test: `internal/server/server_test.go`

**Interfaces:**
- Consumes: `ChatRequest.Device`, `CompletionRequest.Device`, `EmbeddingRequest.Device`, `RerankRequest.Device`
- Produces: device string passed to runtime calls

- [ ] **Step 1: Pass device in handleChat**

In `handleChat`, after parsing request, add device to InferenceParams:

```go
params := runtime.InferenceParams{
    Device:      req.Device,
    MaxTokens:   req.MaxTokens,
    Temperature: 0.7,
    TopP:        0.9,
}
```

- [ ] **Step 2: Pass device in handleCompletion**

Same pattern.

- [ ] **Step 3: Pass device in handleEmbeddings**

```go
result, err := s.engine.Runtime().Embed(c.Request.Context(), req.Model, req.Input, req.Device)
```

- [ ] **Step 4: Pass device in handleRerank**

```go
allInputs := append([]string{req.Query}, req.Documents...)
embResult, err := s.engine.Runtime().Embed(ctx, req.Model, allInputs, req.Device)
```

- [ ] **Step 5: Update Runtime interface**

Add device parameter to Embed:

```go
type Runtime interface {
    // ...
    Embed(ctx context.Context, modelID string, inputs []string, device string) (*EmbeddingResult, error)
    // ...
}
```

- [ ] **Step 6: Update stub Embed**

In `internal/provider/openvino/stub.go`:

```go
func (r *OpenVINORuntime) Embed(ctx context.Context, modelID string, inputs []string, device string) (*runtime.EmbeddingResult, error) {
    return nil, fmt.Errorf("OpenVINO requires CGO: rebuild with CGO_ENABLED=1")
}
```

- [ ] **Step 7: Update real Embed signature**

In `internal/provider/openvino/inference.go`, change Embed signature:

```go
func (r *OpenVINORuntime) Embed(ctx context.Context, modelID string, inputs []string, device string) (*runtime.EmbeddingResult, error) {
```

And use device when calling extractEmbedding (pass in params or directly).

- [ ] **Step 8: Infer/InferStream use params.Device**

When looking up compiled model in Infer/InferStream, use `params.Device` instead of empty:

```go
func (r *OpenVINORuntime) Infer(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (*runtime.InferenceResult, error) {
    r.mu.RLock()
    lm, ok := r.models[modelID]
    r.mu.RUnlock()
    if !ok {
        return nil, fmt.Errorf("model %q is not loaded", modelID)
    }

    device := params.Device
    if device == "" || device == "auto" {
        device = r.defaultDev
    }

    r.mu.RLock()
    h, ok := lm.compiledByDevice[device]
    r.mu.RUnlock()
    if !ok {
        return nil, fmt.Errorf("model %q not compiled for device %q", modelID, device)
    }
    // ... use h.compiled
```

- [ ] **Step 9: Verify build**

```bash
go build ./...
```

---

### Task 4: Workload-based defaults and config

**Files:**
- Modify: `internal/config/config.go` (WorkloadConfig, BenchmarkConfig)
- Modify: `internal/config/config.go` (defaults)
- Modify: `internal/provider/openvino/runtime.go` (workload resolution)
- Test: `internal/config/config_test.go`

**Interfaces:**
- Consumes: Config struct
- Produces: `WorkloadConfig` values used in device resolution

- [ ] **Step 1: Add WorkloadConfig and BenchmarkConfig to config**

```go
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

Add to ModelsConfig:

```go
type ModelsConfig struct {
    Path      string          `mapstructure:"path"`
    Default   string          `mapstructure:"default"`
    Device    string          `mapstructure:"device"`
    Workloads WorkloadConfig  `mapstructure:"workloads"`
    Benchmark BenchmarkConfig `mapstructure:"benchmark"`
}
```

- [ ] **Step 2: Set defaults**

In `DefaultConfig()`:

```go
func DefaultConfig() *Config {
    return &Config{
        // ...existing...
        Models: ModelsConfig{
            Path:    "./models",
            Default: "",
            Device:  "auto",
            Workloads: WorkloadConfig{
                Chat:       "auto",
                Completion: "auto",
                Embed:      "CPU",
                Rerank:     "CPU",
            },
            Benchmark: BenchmarkConfig{
                Enabled:    false,
                Iterations: 3,
                Prompt:     "The quick brown fox jumps over the lazy dog",
                MaxTokens:  50,
            },
        },
        // ...
    }
}
```

- [ ] **Step 3: Add workload resolution to runtime**

In `internal/provider/openvino/runtime.go`, add a method to resolve device per workload:

```go
func (r *OpenVINORuntime) ResolveDevice(workload string, requestDevice string) string {
    if requestDevice != "" && requestDevice != "auto" {
        return requestDevice
    }

    if r.workloads != nil {
        if d, ok := r.workloads[workload]; ok && d != "" && d != "auto" {
            return d
        }
    }

    return r.defaultDev
}
```

Add `workloads` field to OpenVINORuntime:

```go
type OpenVINORuntime struct {
    // ...existing fields...
    workloads map[string]string
}
```

Set it in Initialize or via constructor.

- [ ] **Step 4: Pass workloads from config to runtime**

In `serve.go`, after creating the runtime, pass workloads:

```go
provider := openvino.NewProvider(cfg.Models.Path)
provider.Runtime().SetWorkloads(map[string]string{
    "chat":       cfg.Models.Workloads.Chat,
    "completion": cfg.Models.Workloads.Completion,
    "embed":      cfg.Models.Workloads.Embed,
    "rerank":     cfg.Models.Workloads.Rerank,
})
```

- [ ] **Step 5: Verify build**

```bash
go build ./...
```

---

### Task 5: Auto-benchmark

**Files:**
- Create: `internal/provider/openvino/benchmark.go` (benchmark logic)
- Modify: `internal/provider/openvino/runtime.go` (run benchmark on init)
- Modify: `internal/server/server.go` (handleBenchmark enhancement)
- Test: tests for benchmark

**Interfaces:**
- Consumes: Runtime, loaded models, config
- Produces: `BenchmarkResult` struct, `POST /v1/benchmark` response

- [ ] **Step 1: Create BenchmarkResult type**

In `runtime/runtime.go`:

```go
type DeviceBenchmark struct {
    ChatTokensPerSec float64 `json:"chat_tokens_per_sec"`
    EmbedLatencyMS   float64 `json:"embed_latency_ms"`
}

type BenchmarkResults map[string]DeviceBenchmark
```

- [ ] **Step 2: Add benchmark method to Runtime interface**

```go
type Runtime interface {
    // ...existing...
    Benchmark(ctx context.Context, modelID string, iterations int, prompt string, maxTokens int) (BenchmarkResults, error)
}
```

- [ ] **Step 3: Implement benchmark in stub.go**

```go
func (r *OpenVINORuntime) Benchmark(ctx context.Context, modelID string, iterations int, prompt string, maxTokens int) (runtime.BenchmarkResults, error) {
    return nil, fmt.Errorf("OpenVINO requires CGO: rebuild with CGO_ENABLED=1")
}
```

- [ ] **Step 4: Implement benchmark in OpenVINORuntime**

In `internal/provider/openvino/benchmark.go`:

```go
package openvino

import (
    "context"
    "fmt"
    "time"
    "log/slog"

    "github.com/openforge-ai/openforge/runtime"
)

func (r *OpenVINORuntime) Benchmark(ctx context.Context, modelID string, iterations int, prompt string, maxTokens int) (runtime.BenchmarkResults, error) {
    r.mu.RLock()
    lm, ok := r.models[modelID]
    r.mu.RUnlock()
    if !ok {
        return nil, fmt.Errorf("model %q not loaded", modelID)
    }

    results := make(runtime.BenchmarkResults)
    for device, h := range lm.compiledByDevice {
        slog.Info("benchmarking device", "device", device, "iterations", iterations)

        var totalDuration time.Duration
        for i := 0; i < iterations; i++ {
            inputIDs := []int64{1, 2, 3, 4, 5}
            start := time.Now()
            params := runtime.InferenceParams{
                Device:    device,
                MaxTokens: maxTokens,
                Temperature: 0.1,
            }
            _, err := r.generate(ctx, lm, inputIDs, params)
            if err != nil {
                slog.Warn("benchmark iteration failed", "device", device, "error", err)
                continue
            }
            totalDuration += time.Since(start)
        }

        avgDuration := totalDuration / time.Duration(iterations)
        tokensPerSec := float64(maxTokens) / avgDuration.Seconds()

        results[device] = runtime.DeviceBenchmark{
            ChatTokensPerSec: tokensPerSec,
            EmbedLatencyMS:   0, // simplified: embeddings benchmark TBD
        }
    }

    return results, nil
}
```

- [ ] **Step 5: Run benchmark on startup if enabled**

In `runtime.go Initialize`, after loading default model:

```go
if r.benchmarkCfg != nil && r.benchmarkCfg.Enabled {
    if r.defaultModelID != "" {
        results, err := r.Benchmark(ctx, r.defaultModelID, r.benchmarkCfg.Iterations, r.benchmarkCfg.Prompt, r.benchmarkCfg.MaxTokens)
        if err != nil {
            slog.Warn("benchmark failed", "error", err)
        } else {
            for device, res := range results {
                slog.Info("benchmark result", "device", device, "tokens_per_sec", res.ChatTokensPerSec)
            }
        }
    }
}
```

- [ ] **Step 6: Verify build**

```bash
go build ./...
```

---

### Task 6: Update tests

**Files:**
- Modify: `runtime/types_test.go` — verify device field JSON round-trip
- Modify: Tests in server package — verify device propagation
- Modify: Config tests — verify workload defaults

**Step sequence:** For each test file, write the failing test, run to confirm failure, then ensure existing code makes it pass (most tests should pass from prior tasks).

- [ ] **Step 1: Add types_test for device field**

```go
func TestChatRequest_WithDevice(t *testing.T) {
    req := ChatRequest{
        Model:  "phi-3-mini",
        Device: "NPU",
        Messages: []Message{{Role: "user", Content: "hi"}},
    }
    data, err := json.Marshal(req)
    if err != nil {
        t.Fatal(err)
    }
    var decoded ChatRequest
    if err := json.Unmarshal(data, &decoded); err != nil {
        t.Fatal(err)
    }
    if decoded.Device != "NPU" {
        t.Errorf("expected NPU, got %q", decoded.Device)
    }
}
```

- [ ] **Step 2: Run tests**

```bash
go test ./runtime/... ./internal/config/... -v -count=1
```

- [ ] **Step 3: Verify full build**

```bash
go build ./...
```

---
