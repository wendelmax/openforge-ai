# Providers Guide

> **⚠️ Aspirational:** The Plugin API for third-party providers is not yet implemented. Only the native OpenVINO provider is available.

Providers abstract the inference backend from the engine. Currently, OpenForge supports one native provider (OpenVINO). The Plugin API enables third-party implementations.

## Provider Interface

```go
type Provider interface {
    Name() string
    Runtime() Runtime
    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

## Runtime Interface

```go
type Runtime interface {
    ListDevices(ctx context.Context) ([]Device, error)
    LoadModel(ctx context.Context, modelID, path, device string) error
    UnloadModel(ctx context.Context, modelID string) error
    ListModels(ctx context.Context) ([]ModelInfo, error)
    Infer(ctx context.Context, modelID string, prompt string, params InferenceParams) (*InferenceResult, error)
    InferStream(ctx context.Context, modelID string, prompt string, params InferenceParams) (<-chan string, error)
    Embed(ctx context.Context, modelID string, inputs []string) (*EmbeddingResult, error)
    Close(ctx context.Context) error
}
```

## OpenVINO Provider (Default)

The built-in provider uses CGO to interface with the OpenVINO C API.

### Capabilities

| Feature | Supported |
|---------|:---------:|
| CPU inference | ✅ |
| GPU inference (Iris, Arc) | ✅ |
| NPU inference (AI Boost) | ✅ |
| Auto device selection | ✅ |
| Async inference | ✅ |
| Streaming (SSE) | ✅ |
| Embedding extraction | ✅ |
| Reranking | ✅ |
| INT4/FP16/FP32 | ✅ |

### Configuration

```yaml
models:
  path: "./models"       # Model directory
  default: "phi-3-mini"  # Default model (optional)
  device: "auto"         # "auto", "CPU", "GPU.0", "NPU"
```

### Device Priority

When `device: auto`:

1. **GPU** (if available) — best throughput for LLMs
2. **NPU** (if available) — best efficiency for embeddings
3. **CPU** — universal fallback

## Creating a Custom Provider (Plugin API)

### Step 1: Define Plugin

```go
// my-provider/main.go
package main

import (
    "context"
    "github.com/openforge-ai/openforge/runtime"
)

type MyRuntime struct{}

func (r *MyRuntime) ListDevices(ctx context.Context) ([]runtime.Device, error) {
    return []runtime.Device{
        {ID: "my-device", Name: "Custom Device", Type: runtime.DeviceCPU, Available: true},
    }, nil
}

func (r *MyRuntime) LoadModel(ctx context.Context, modelID, path, device string) error {
    // Custom model loading
    return nil
}

// Implement remaining Runtime methods...

type MyProvider struct {
    runtime *MyRuntime
}

func (p *MyProvider) Name() string { return "my-provider" }
func (p *MyProvider) Runtime() runtime.Runtime { return p.runtime }
func (p *MyProvider) Initialize(ctx context.Context) error { return nil }
func (p *MyProvider) Shutdown(ctx context.Context) error { return nil }

// Export as plugin
var Plugin = MyProvider{}
```

### Step 2: Build as Shared Object

```bash
go build -buildmode=plugin -o my-provider.so my-provider/main.go
```

### Step 3: Install

```bash
cp my-provider.so ~/.openforge/plugins/
```

### Step 4: Configure

```yaml
plugins:
  - name: my-provider
    path: ~/.openforge/plugins/my-provider.so
```

## Provider Checklist

| Requirement | Importance |
|-------------|:----------:|
| Implement all `Runtime` methods | Required |
| Thread-safe operations | Required |
| Graceful shutdown | Required |
| Report available devices | Required |
| Support streaming | Recommended |
| Support embeddings | Recommended |
| Version metadata | Recommended |
| Benchmarks | Community expectation |
