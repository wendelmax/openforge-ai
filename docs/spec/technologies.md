# Technology Stack

## Primary

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.23+ | Linguagem principal |
| OpenVINO | 2025.x | Runtime de inferência IA |
| Gin | v1.10+ | HTTP framework |
| SQLite | 3.x | Cache vetorial e persistência |
| BGE Small EN | v1.5 | Modelo de embeddings padrão |

## Build & Tooling

| Tool | Purpose |
|------|---------|
| Taskfile | Automação de build (makefile alternative) |
| golangci-lint | Linter |
| mockgen | Geração de mocks |
| go-sqlite3 | Driver SQLite |
| testcontainers-go | Testes de integração |
| pprof | Profiling |

## Observability

| Tool | Purpose |
|------|---------|
| OpenTelemetry | Tracing, métricas, logs (opcional) |
| slog | Logging estruturado padrão |
| expvar | Métricas de runtime |

## Transport

| Protocol | Purpose |
|----------|---------|
| HTTP/1.1 | API REST |
| SSE | Streaming de respostas |
| Unix Socket | IPC local (futuro) |

## Format

| Format | Purpose |
|--------|---------|
| JSON | API payloads |
| YAML | Configuração |
| OpenVINO IR | Model weights (XML + BIN) |
| SQLite | Cache e persistência |
| Markdown | Documentação |

## CI/CD

| Tool | Purpose |
|------|---------|
| GitHub Actions | CI/CD pipeline |
| GoReleaser | Build e release automatizado |
| Docker | Containerização |
| Codecov | Cobertura de testes |

## Development

| Tool | Purpose |
|------|---------|
| VS Code | IDE primária |
| GoLand | IDE alternativa |
| Delve | Debugger |
| Air | Live reload |
| Vulcand/static | Static file server para modelos |

## Compatibility Targets

| Platform | Support |
|----------|---------|
| Linux (x86_64) | Tier 1 — Full support |
| Windows (x86_64) | Tier 2 — Tested |
| macOS (ARM64) | Tier 3 — Experimental |
| Linux (ARM64) | Tier 3 — Experimental |

## Hardware Targets

| Hardware | Support |
|----------|---------|
| Intel CPU (Core, Xeon) | Full |
| Intel GPU (Iris, Arc) | Full |
| Intel NPU (AI Boost) | Full |
| AMD CPU | Functional (CPU only) |
| Apple Silicon | Experimental (CPU only) |

## Package Dependencies (Go)

```
github.com/gin-gonic/gin
github.com/mattn/go-sqlite3
github.com/stretchr/testify
github.com/fsnotify/fsnotify
github.com/spf13/viper
github.com/google/uuid
go.opentelemetry.io/otel
go.opentelemetry.io/otel/exporters/otlp
github.com/vektra/mockery/v2 (dev)
github.com/golangci/golangci-lint (dev)
```
