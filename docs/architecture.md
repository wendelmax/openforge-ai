# Architecture

## High-Level Overview

```mermaid
graph TB
    OC["OpenCode / Superpowers<br/>(IDE, CLI, VS Code, IntelliJ)"]
    OC -->|"HTTP REST (OpenAI-compatible)"| API

    subgraph API_Layer["OpenForge API Layer (Gin HTTP)"]
        API["/v1/chat  /v1/completion  /v1/embeddings<br/>/v1/rerank  /v1/models  /v1/devices"]
    end

    API --> Engine

    subgraph Engine_Layer["AI Engine"]
        Engine["Session Manager | Context Manager<br/>Cache Manager | Skill Executor<br/>Pipeline Engine | Telemetry Collector"]
    end

    Engine --> Provider

    subgraph Provider_Layer["Provider Layer"]
        subgraph OV["OpenVINO Provider"]
            OV1["Model Lifecycle | Device Detection<br/>Inference | Embedding"]
        end
        subgraph Plugin["Plugin API (future)"]
            P1["Custom providers via .so/.dll"]
        end
    end

    Provider --> OVR

    subgraph OVR_Layer["OpenVINO Runtime"]
        OVR["Core | Inference Engine | Model Loading<br/>Device Discovery | Tensor Management"]
    end

    OVR --> CPU
    OVR --> GPU
    OVR --> NPU

    CPU["CPU (x86_64)"]
    GPU["GPU (Iris/Arc)"]
    NPU["NPU (AI Boost)"]
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

```mermaid
graph TD
    API["API Layer"] --> Engine["Engine"] --> Provider["Provider"] --> OVR["OpenVINO Runtime"]
    API -.-> Log["Logging (slog) / Metrics (OpenTelemetry)"]
    Engine -.-> SQLite["SQLite (cache, sessions, embeddings)"]
    Provider -.-> Drivers["Device drivers (GPU, NPU)"]
```

Rules:
- Dependencies point **inward** (API → Engine → Provider → Runtime)
- No layer imports from layers above it
- All inter-layer communication through Go interfaces

## Communication Patterns

| Layer | Protocol | Format |
|-------|----------|--------|
| OpenCode ↔ API | HTTP/1.1 | JSON |
| API ↔ Engine | Go function call | Go structs |
| Engine ↔ Provider | Go interface | Go structs |
| Provider ↔ OpenVINO | CGO / OpenVINO C API | C buffers |

## Cross-Cutting Concerns

- **Logging**: structured logs via slog (standard library)
- **Config**: hierarchical merge per layer
- **Errors**: typed errors with wrapped context
- **Metrics**: counters and histograms (future: OpenTelemetry)
- **Tracing**: distributed spans for critical operations

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Single binary | Easy deployment, no runtime dependencies |
| CGO for OpenVINO | Direct C API access, minimal overhead |
| Interfaces at boundaries | Testable, swappable implementations |
| Async everywhere | Non-blocking inference, streaming support |
| Cache-first design | Embedding cache reduces latency 10x+ |

## Data Flow: Chat Request

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Engine
    participant Cache
    participant Provider
    participant OpenVINO

    Client->>API: HTTP POST /v1/chat
    API->>API: validate request body
    API->>Engine: retrieve/create session
    Engine->>Engine: append user message to context
    Engine->>Cache: check response cache (hash = model + messages + params)
    alt Cache HIT
        Cache-->>Engine: return cached response
        Engine-->>API: cached response
        API-->>Client: response
    else Cache MISS
        Engine->>Provider: Provider.Infer()
        Provider->>Provider: compile prompt from messages
        Provider->>Provider: tokenize input
        Provider->>OpenVINO: inference loop (autoregressive generation)
        OpenVINO-->>Provider: output tokens
        Provider->>Provider: detokenize output
        Provider-->>Engine: response
        Engine->>Cache: cache response
        Engine-->>API: response (streaming or complete)
        API-->>Client: response
    end
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

```mermaid
graph TD
    A["SIGTERM/SIGINT"] --> B["Server stops accepting connections"]
    B --> C["In-flight requests complete (with timeout)"]
    C --> D["Sessions persist to disk"]
    D --> E["Models unload from devices"]
    E --> F["Cache flushed to SQLite"]
    F --> G["Process exits"]
```
