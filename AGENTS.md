# OpenForge — AI Agent Guide

## Project overview

OpenForge is a **100% local AI coding agent** written in Go (1.25+). It reads, edits, and executes code via LLM-driven tools — no cloud required. It exposes an OpenAI-compatible API, a Bubble Tea TUI, and a CLI (`openforge`). Primary inference runs through OpenVINO (CGO), with HTTP-based fallback providers (Ollama, llama.cpp, vLLM, LM Studio).

Module path: `github.com/openforge-ai/openforge`

## Essential commands

Use `task` (go-task runner) for all development workflows:

| Command | What it does |
|---|---|
| `task test` | `go test ./... -v -count=1` |
| `task test:cover` | Tests with coverage → `coverage.out` + `coverage.html` |
| `task test:race` | Tests with race detector |
| `task lint` | `golangci-lint run ./...` |
| `task vet` | `go vet ./...` |
| `task fmt` | `go fmt ./...` |
| `task build` | Build stub binary (`CGO_ENABLED=0`) to `build/openforge` |
| `task run` | `go run ./cmd/server --config config.example.yaml` |
| `task benchmark` | `go test ./... -bench=. -benchmem` |
| `task swagger` | Generate OpenAPI spec via `swag` |

**Windows smoke test**: `hack/smoke-test.bat`

**CI**: `.github/workflows/ci.yml` runs lint → test (CGO=0 + CGO=1) → multi-platform build (linux/windows/darwin × amd64/arm64 stub) → CGO Docker build → benchmarks (main only). Go version in CI is 1.26.

## Architecture

### Dependency layers (top → bottom)

```
cmd/openforge, cmd/server         (entry points)
    ↓
internal/tui, internal/server     (presentation)
    ↓
internal/agent, internal/engine   (orchestration)
    ↓
internal/pm                       (provider manager/adapter layer)
    ↓
internal/provider/openvino,       (runtime implementations)
internal/pm/providers/*           (HTTP-based providers: ollama, openai_compat)
    ↓
runtime/                          (shared domain types + Runtime interface)
```

**Rule**: dependencies point inward. No layer imports from layers above it. All inter-layer communication goes through Go interfaces.

### Key interfaces

- **`runtime.Runtime`** (`runtime/runtime.go`) — the foundational interface: `Infer`, `InferStream`, `Embed`, `LoadModel`, `UnloadModel`, `ListDevices`, `Benchmark`, `DeviceForWorkload`
- **`pm.Provider`** (`internal/pm/types.go`) — the logical provider interface agents use: `Chat`, `ChatStream`, `Embed`, `ListModels`, `LoadModel`, `Start`, `Stop`
- **`tool.Tool`** (`internal/tool/tool.go`) — agent tool contract: `Name()`, `Description()`, `Run(ctx, args)`, `InputSchema()`
- **`skill.ExecutorInterface`** (`internal/skill/skill.go`) — skill pipeline execution
- **`engine.SessionStore`** (`internal/engine/`) — session persistence (Memory/File/SQLite backends)
- **`permission.Store`** (`internal/permission/`) — permission rule persistence

### Double provider abstraction

There are **two** provider interfaces — this is intentional, not accidental:

1. `runtime.Runtime` — physical inference runtime (OpenVINO C API). Low-level.
2. `pm.Provider` — logical provider (what agents use). Higher-level, includes auto-discovery, health checks, Chat/ChatStream with tool-def support.

`pm.OpenVINOAdapter` adapts `openvino.OpenVINORuntime` → `pm.Provider`. HTTP providers (`OllamaProvider`, `OpenAICompatProvider`) implement `pm.Provider` directly without the `runtime.Runtime` layer.

### Package map

| Package | Role | Imports from |
|---|---|---|
| `runtime/` | Shared domain types (`Message`, `Device`, `Runtime` interface) | nothing internal |
| `internal/config/` | Viper-based config loading, env override (`OPENFORGE_` prefix) | nothing internal |
| `internal/cache/` | In-memory cache with SQLite persistence | nothing internal |
| `internal/modelzoo/` | Model catalog metadata | nothing internal |
| `internal/hooks/` | PreToolUse/PostToolUse hook engine (runs external scripts) | nothing internal |
| `internal/lsp/` | LSP client manager (gopls, tsserver, rust-analyzer, pyright) | nothing internal |
| `internal/permission/` | Tool-usage permission manager (ask/allow/deny) | nothing internal |
| `internal/skills/` | SKILL.md discovery + injection | nothing internal (leaf) |
| `internal/mcp/` | MCP client + registry (stdio subprocess transport) | `config`, `permission` |
| `internal/skill/` | Skill pipeline executor (prompt → embed → rerank → format → tool) | `runtime` |
| `internal/tool/` | Tool registry + built-in tools (bash, view, edit, grep, etc.) | `mcp` (MCPAdapter) |
| `internal/pm/` | Provider manager: auto-detect, health-check, chain routing | `runtime`, `provider/openvino`, `pm/providers/*` |
| `internal/agent/` | Agent loop with tool-calling (max 10 iterations) | `pm`, `tool` |
| `internal/engine/` | Session/context manager + inference orchestration | `runtime`, `skill`, `config` |
| `internal/tui/` | Bubble Tea terminal UI | `agent`, `pm`, `tool` |
| `internal/server/` | Gin HTTP server (OpenAI-compatible endpoints) | `runtime`, `engine`, `config`, `pm` |

### Entry points

- **`cmd/openforge/`** — Cobra CLI. Subcommands: `serve`, `provider list/detect/install/guide`, `model list`, `devices`, `benchmark`, `skill discover/inspect`, `version`, default=TUI.
- **`cmd/server/`** — standalone HTTP server binary (lighter wiring than `openforge serve`).

Both wire up `openvino.Provider` → `ProviderManager` → `Engine` → `Server`, but `cmd/server` also sets up MCP and permission systems.

## CGO & stub mode

OpenVINO is the native inference backend, linked via CGO (`// #cgo LDFLAGS: -lopenvino_c`). This makes cross-compilation impossible with CGO enabled.

**Stub mode** (`CGO_ENABLED=0`): builds without OpenVINO CGO. The OpenVINO provider returns stub/no-op implementations. This is the default for local development — use `task build` or `task test`.

**CGO builds** require the Docker builder image (`build/Dockerfile.builder`) which bundles OpenVINO SDK 2026.2.1. Use `task dist:linux-cgo` for release builds.

**When writing tests**: always test with `CGO_ENABLED=0` first. CGO-enabled tests run in CI but require OpenVINO libraries installed. Stub tests should never panic or fail due to missing CGO symbols — use build tags or conditional checks.

## Testing conventions

- Unit tests: `_test.go` alongside source, same package (white-box)
- Integration tests: `//go:build integration` build tag
- Mocks: `internal/mocks/` (generated via `mockery`)
- Table-driven tests preferred for multi-scenario coverage
- Naming: `TestFuncName_Scenario_ExpectedBehavior`
- CI target: 90% coverage minimum

## Code style & gotchas

### Imports
Three groups, blank-line separated: stdlib → third-party → internal. Example:
```go
import (
    "context"
    "fmt"

    "github.com/spf13/viper"

    "github.com/openforge-ai/openforge/runtime"
)
```

### Naming
- Avoid abbreviations except: `ctx`, `wg`, `mu`, `err`
- Packages: lowercase, singular (`model.go` not `models.go`)
- One primary type per file
- `New()` for constructors, `NewRegistry()`, `NewManager()` for composed types

### Patterns observed
- **Concrete structs dominate**: only 7 interfaces across 16 `internal/` packages. "Accept interfaces, return structs" style.
- **Every `pm.Provider` implementation** returns `(*ProviderHealth, nil)` on error — errors are embedded in `ProviderHealth.Error`, never returned as Go `error`. This is intentional: allows the provider manager to collect health from all providers without short-circuiting.
- **Config is Viper with env override**: config keys use `OPENFORGE_` prefix (e.g., `OPENFORGE_SERVER_PORT`). See `.env.example`.
- **No `init()` functions, no globals, no singletons** — everything is explicit DI via constructors.
- **`runtime/` is NOT under `internal/`** — it is the shared domain types package used across the entire codebase.

### gopls warnings (non-blocking)
The project has ~90 gopls hints (not errors): `interface{}` → `any`, `for` loop modernization, `strings.Cut`, `slices.Contains`. These are style modernization hints, not bugs. Don't bulk-fix them unless asked.

## ADRs

Architectural decisions live in `docs/adr/`. Key decisions:
- **ADR-0001**: OpenVINO as the only native inference runtime
- **ADR-0002**: Go as the primary language (CGO for OpenVINO bindings)
- **ADR-0003**: Plugin system design
- **ADR-0004**: Agent Skills pipeline architecture
- **ADR-0005**: RAG architecture
- **ADR-0006**: OpenAPI 3.0 + swaggo for API docs

**Always read relevant ADRs before making architectural changes.**

## Config file

`openforge.yaml` — auto-generated, YAML. Provider chain order, workload routing, server settings, logging. Environment variable overrides use `OPENFORGE_` prefix (uppercase, underscores for nesting).

## Skills & hooks

- **Skills** are YAML-defined pipelines (`skills/` directory, SKILL.md format). Each skill has steps of type: `prompt`, `embed`, `rerank`, `format`, `tool`. Discovered via `internal/skills/`.
- **Hooks** run external scripts before/after tool use. Configured in the config file under `hooks:`. Events: `PreToolUse`, `PostToolUse`. Hook scripts receive tool name and args via stdin as JSON; return `{"allowed": true/false}` + optional message.

## Tool system

9 built-in tools registered in `tool.DefaultRegistry()`: bash, view, write, edit, grep, glob, ls, todos, fetch. MCP tools bridge via `tool.MCPAdapter` which adapts an `mcp.Client` to the `tool.Tool` interface. Tool calls have a hard limit of 10 iterations per agent run (`maxToolIterations` in `agent.go:15`).
