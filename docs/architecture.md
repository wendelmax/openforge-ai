# Architecture

## High-Level Overview

```
┌──────────────────────────────────────────────┐
│              OpenCode / Superpowers            │
│             (IDE, CLI, VS Code, IntelliJ)      │
└────────────────────┬─────────────────────────┘
                     │ HTTP REST (OpenAI-compatible)
                     ▼
┌──────────────────────────────────────────────┐
│              OpenForge API Layer               │
│                   (Gin HTTP)                   │
│  /v1/chat  /v1/completion  /v1/embeddings     │
│  /v1/rerank  /v1/models  /v1/devices          │
└────────────────────┬─────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────┐
│                 AI Engine                      │
│  Session Manager │ Context Manager            │
│  Cache Manager   │ Skill Executor             │
│  Pipeline Engine │ Telemetry Collector        │
└────────────────────┬─────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────┐
│               Provider Layer                   │
│  ┌─────────────────────────────────────────┐  │
│  │         OpenVINO Provider                │  │
│  │  Model Lifecycle │ Device Detection     │  │
│  │  Inference       │ Embedding            │  │
│  └─────────────────────────────────────────┘  │
│  ┌─────────────────────────────────────────┐  │
│  │  Plugin API (future)                    │  │
│  │  Custom providers via .so/.dll          │  │
│  └─────────────────────────────────────────┘  │
└────────────────────┬─────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────┐
│              OpenVINO Runtime                  │
│  Core │ Inference Engine │ Model Loading      │
│  Device Discovery │ Tensor Management         │
└────────────────────┬─────────────────────────┘
                     │
         ┌───────────┼───────────┐
         ▼           ▼           ▼
       CPU         GPU         NPU
    (x86_64)   (Iris/Arc)   (AI Boost)
```

## Layer Responsibilities

### API Layer
- Exposes OpenAI-compatible REST endpoints
- Handles authentication, rate limiting, CORS
- Validates requests, serializes responses
- Serves Swagger UI at `/docs`

### AI Engine
- Orchestrates inference flows (chat, completion, embedding, rerank)
- Manages conversation sessions with context windowing
- Executes Skills as multi-step pipelines
- Caches embeddings and responses (memory + SQLite)
- Collects metrics and traces for observability

### Provider Layer
- Abstracts hardware and runtime details from the Engine
- Implements model lifecycle (load, unload, reload)
- Detects and selects optimal devices automatically
- Plugin API for third-party providers

### OpenVINO Runtime
- Loads OpenVINO IR models (.xml + .bin)
- Compiles models for target devices
- Executes synchronous and asynchronous inference
- Manages device memory and tensor resources

## Dependency Flow

```
API Layer ──→ Engine ──→ Provider ──→ OpenVINO Runtime
    │            │            │
    │            │            └── Device drivers (GPU, NPU)
    │            │
    │            └── SQLite (cache, sessions, embeddings)
    │
    └── Logging (slog) / Metrics (OpenTelemetry)
```

Rules:
- Dependencies point **inward** (API → Engine → Provider → Runtime)
- No layer imports from layers above it
- All inter-layer communication through Go interfaces

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Single binary | Easy deployment, no runtime dependencies |
| CGO for OpenVINO | Direct C API access, minimal overhead |
| Interfaces at boundaries | Testable, swappable implementations |
| Async everywhere | Non-blocking inference, streaming support |
| Cache-first design | Embedding cache reduces latency 10x+ |

## Data Flow: Chat Request

```
1. HTTP POST /v1/chat
2. API validates request body
3. Engine retrieves/creates session
4. Engine appends user message to context
5. Engine checks response cache (hash = model + messages + params)
   ├── Cache HIT  → return cached response immediately
   └── Cache MISS → continue
6. Engine calls Provider.Infer()
7. Provider compiles prompt from messages
8. Provider tokenizes input
9. OpenVINO runs inference loop (autoregressive generation)
10. Provider detokenizes output
11. Engine caches response
12. Engine returns response (streaming or complete)
```

## Thread Safety

All components are designed for concurrent access:

| Component | Strategy |
|-----------|----------|
| Sessions | `sync.RWMutex` per session map |
| Models | Reference-counted, write-once |
| Cache | Sharded mutexes per hash bucket |
| Provider | Single-threaded per model + queue |
| Server | Gin handles concurrency per request |

## Graceful Shutdown

```
SIGTERM/SIGINT
  → Server stops accepting connections
  → In-flight requests complete (with timeout)
  → Sessions persist to disk
  → Models unload from devices
  → Cache flushed to SQLite
  → Process exits
```
