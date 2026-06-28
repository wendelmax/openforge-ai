# Roadmap

Public roadmap for OpenForge development.

## Legend

| Icon | Meaning |
|:----:|:--------|
| ✅ | Done |
| 🔧 | In progress |
| 📋 | Planned |
| 💡 | Proposed |

## v0.1 — Runtime Foundation (Current Sprint)

### Core
- [ ] ✅ Go module and project structure
- [ ] ✅ OpenVINO CGO bindings
- [ ] ✅ Core, Model, CompiledModel wrappers
- [ ] 🔧 Model loading from disk
- [ ] 🔧 Model unloading
- [ ] 📋 Automatic device detection (CPU, GPU, NPU)
- [ ] 📋 Device prioritization (GPU > NPU > CPU)

### API
- [ ] 📋 HTTP server with Gin
- [ ] 📋 POST /v1/health
- [ ] 📋 POST /v1/chat
- [ ] 📋 POST /v1/models
- [ ] 📋 POST /v1/devices

### Docs
- [ ] ✅ Spec documents (12 of 12)
- [ ] ✅ ADRs (6 of 6)
- [ ] ✅ Architecture documentation
- [ ] ✅ Installation guide
- [ ] ✅ Quickstart

## v0.2 — Inference & Embeddings

### Inference
- [ ] 📋 Synchronous inference
- [ ] 📋 Streaming inference (SSE)
- [ ] 📋 Token generation loop
- [ ] 📋 Top-k, top-p, temperature sampling

### Embeddings
- [ ] 📋 BGE Small integration
- [ ] 📋 Embedding extraction (mean pooling)
- [ ] 📋 Embedding cache (memory)
- [ ] 📋 POST /v1/embeddings

### Provider
- [ ] 📋 OpenVINO provider implementation
- [ ] 📋 Device fallback (GPU → NPU → CPU)
- [ ] 📋 Provider configuration

## v0.3 — Skills

### Skills Engine
- [ ] 📋 YAML skill loader
- [ ] 📋 Pipeline executor
- [ ] 📋 Step types: prompt, embed, rerank, format
- [ ] 📋 Template variables and expressions

### Skills Catalog
- [ ] 📋 Go skill
- [ ] 📋 Java/Spring skill
- [ ] 📋 Python skill
- [ ] 📋 Docker skill
- [ ] 📋 Code review skill

## v0.4 — Sessions & Cache

### Sessions
- [ ] 📋 Session management
- [ ] 📋 Context windowing
- [ ] 📋 Session persistence

### Cache
- [ ] 📋 SQLite embedding cache
- [ ] 📋 SQLite response cache
- [ ] 📋 Cache invalidation
- [ ] 📋 Configurable cache backend

## v0.5 — RAG

- [ ] 📋 Chunking (recursive text splitter)
- [ ] 📋 Vector storage (SQLite + vec extension)
- [ ] 📋 Retrieval pipeline
- [ ] 📋 Reranking
- [ ] 📋 POST /v1/rerank

## v0.6 — Observability

- [ ] 📋 Structured logging
- [ ] 📋 Prometheus metrics endpoint
- [ ] 📋 OpenTelemetry traces
- [ ] 📋 pprof endpoints
- [ ] 📋 Health check with model status

## v0.7 — Plugin System

- [ ] 📋 Plugin interface
- [ ] 📋 Shared object loader (.so/.dll)
- [ ] 📋 Plugin lifecycle management
- [ ] 📋 Example plugin

## v0.8 — Benchmark & Release

- [ ] 📋 Benchmark CLI
- [ ] 📋 Benchmark reporting
- [ ] 📋 GoReleaser config
- [ ] 📋 Docker multi-stage build
- [ ] 📋 Homebrew formula
- [ ] 📋 Release v0.8.0

## v1.0 — Stable

- [ ] 📋 API stability guarantee
- [ ] 📋 90%+ test coverage
- [ ] 📋 Cross-platform CI
- [ ] 📋 Performance regression tests
- [ ] 📋 Security audit
- [ ] 📋 Release v1.0.0

## Future

- [ ] 💡 VS Code extension
- [ ] 💡 IntelliJ plugin
- [ ] 💡 Multi-model serving
- [ ] 💡 Vision (YOLO, ResNet)
- [ ] 💡 Audio (Whisper)
- [ ] 💡 Web UI dashboard
- [ ] 💡 WASM plugins
- [ ] 💡 Model download service

---

*Last updated: June 2026*

This roadmap is a living document. Priorities may shift based on community feedback.
