# OpenForge

**AI Runtime for Developers** — 100% local, 100% OpenVINO. CPU, GPU, NPU.

```bash
openforge                # TUI interativo (chat com modelos locais)
openforge serve          # HTTP server (API compatível com OpenAI)
openforge model list     # Lista modelos disponíveis
```

## Features

- **TUI interativo** — chat com streaming token a token, status bar em tempo real
- **Smart device selection** — CPU/GPU/NPU com fallback automático, override por request
- **Workload defaults** — chat → GPU, embedding → CPU, configurável por workload
- **OpenAI-compatible API** — `/v1/chat`, `/v1/completions`, `/v1/embeddings`, `/v1/rerank`
- **OpenAPI spec** — `GET /openapi.json` com schema completo
- **Auto-benchmark** — benchmark de performance no startup, tok/s por device
- **Embedding cache** — cache in-memory com TTL para embeddings repetidos
- **CGO + Stub** — compila com ou sem OpenVINO (stub mode para desenvolvimento)
- **Multi-plataforma** — Linux amd64 (CGO), Windows amd64 (CGO nativo), arm64/darwin (stub)

## Quickstart

```bash
# Stub mode (sem OpenVINO — desenvolvimento)
go build -o openforge ./cmd/openforge

# TUI interativo
./openforge

# HTTP server
./openforge serve

# Com modelo real (requer CGO + OpenVINO)
CGO_ENABLED=1 go build -o openforge ./cmd/openforge
./openforge serve --model phi-3-mini

# Chat via API
curl -X POST http://localhost:9090/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"model":"phi-3-mini","messages":[{"role":"user","content":"hello"}]}'

# Especificar dispositivo
curl -X POST http://localhost:9090/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"model":"phi-3-mini","device":"NPU","messages":[{"role":"user","content":"hello"}]}'
```

## TUI

```
◆ OpenForge ─────────── model: phi-3-mini │ device: GPU ─── /help • Ctrl+C exit
────────────────────────────────────────────────────────────────────────────────
┃ You
  explain quantum computing

┃ OpenForge
  Quantum computing uses qubits instead of bits...
  (streaming token by token)

────────────────────────────────────────────────────────────────────────────────
❯ /help                                                                        
GPU │ model: phi-3-mini │ 45.2 tok/s ─────────────────── /help • Ctrl+C exit
```

| Comando | Atalho | Descrição |
|---------|--------|-----------|
| `/model <name>` | `/m` | Carrega e seleciona modelo |
| `/model` | `/m` | Lista modelos disponíveis |
| `/device <name>` | `/d` | Troca dispositivo (Tab cicla) |
| `/device` | `/d` | Lista dispositivos |
| `/clear` | `/c` | Limpa chat |
| `/exit` | `/q` | Sai |
| `/help` | `/h` | Ajuda |

## CLI

| Command | Description |
|---------|-------------|
| `openforge` | Launch interactive TUI |
| `openforge serve` | Start HTTP API server |
| `openforge model list` | List available models |
| `openforge model load <id>` | Load model into memory |
| `openforge model unload <id>` | Unload model from memory |
| `openforge devices` | List available hardware devices |
| `openforge benchmark` | Run performance benchmarks |
| `openforge version` | Print version |

## Smart Device Selection

```
Resolution order:  request.device > workload config > auto-benchmark > static priority
Priority:          GPU > NPU > CPU
```

Configure em `config.yaml`:

```yaml
devices:
  default: "auto"
  chat: "GPU"
  embedding: "CPU"
  rerank: "CPU"

benchmark:
  enabled: true
  iterations: 3
```

## Build

```bash
# Stub (qualquer plataforma)
CGO_ENABLED=0 go build -o openforge ./cmd/openforge

# CGO + OpenVINO (Linux)
CGO_ENABLED=1 go build -o openforge ./cmd/openforge

# CGO + OpenVINO (Windows — PowerShell)
$env:CGO_ENABLED=1
$env:CGO_CFLAGS="-IC:\Users\...\openvino_env\Lib\site-packages\openvino\libs"
$env:CGO_LDFLAGS="-LC:\Users\...\openvino_env\Lib\site-packages\openvino\libs -lopenvino_c"
go build -o openforge.exe ./cmd/openforge
```

### Release builds (via Goreleaser + GitHub Actions)

| Platform | Type | Archive |
|----------|------|---------|
| Linux amd64 | CGO + OpenVINO | `.tar.gz` com `.so` |
| Linux arm64 | Stub | `.tar.gz` |
| Windows amd64 | CGO + OpenVINO | `.zip` com `.dll` |
| Windows arm64 | Stub | `.tar.gz` |
| Darwin arm64 | Stub | `.tar.gz` |
| Docker | CGO + OpenVINO | `ghcr.io/openforge-ai/openforge` |

## Architecture

```
┌──────────────────────────────────────────────────┐
│                    openforge                      │
│  ┌──────────┐  ┌──────────┐  ┌────────────────┐ │
│  │  TUI     │  │  HTTP    │  │  CLI           │ │
│  │ bubbletea│  │  Gin     │  │  cobra         │ │
│  └────┬─────┘  └────┬─────┘  └───────┬────────┘ │
│       └─────────────┼────────────────┘           │
│                     ▼                            │
│  ┌──────────────────────────────────────────┐    │
│  │            OpenVINO Runtime              │    │
│  │  LoadModel  Infer  Embed  Benchmark      │    │
│  │  DeviceForWorkload  selectBestDevice     │    │
│  └──────────────┬───────────────────────────┘    │
│                 ▼                                │
│  ┌──────────────────────────────────────────┐    │
│  │  OpenVINO C API (CGO)                    │    │
│  │  CPU · GPU · NPU                         │    │
│  └──────────────────────────────────────────┘    │
└──────────────────────────────────────────────────┘
```

## Development

```bash
# Prerequisites: Go 1.26+, Task
task build       # Build stub binary
task test        # Run all tests (CGO_ENABLED=0)
task test:cover  # Coverage report
task vet         # go vet
task fmt         # go fmt
task dist        # Build all distribution binaries (CGO + stub)
task dist:linux-cgo  # CGO Linux via Docker builder
```

Requires OpenVINO Runtime for CGO builds. Non-CGO stubs allow compilation without OpenVINO.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/openapi.json` | OpenAPI 3.0 specification |
| POST | `/v1/chat` | Chat completion (streaming via SSE) |
| POST | `/v1/completion` | Text completion |
| POST | `/v1/embeddings` | Generate embeddings |
| POST | `/v1/rerank` | Rerank documents |
| GET | `/v1/models` | List models |
| GET | `/v1/devices` | List devices |
| POST | `/v1/model/load` | Load model |
| POST | `/v1/model/unload` | Unload model |
| POST | `/v1/benchmark` | Run benchmark |

## Docs

- [Architecture](spec/architecture.md)
- [API Reference](docs/api-reference.md)
- [Smart Device Selection](docs/superpowers/specs/2026-06-26-smart-device-selection-design.md)
- [Architectural Decisions](docs/adr/)

## License

Apache 2.0
