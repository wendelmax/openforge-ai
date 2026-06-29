# OpenForge

**AI Coding Agent — 100% Local, 100% Open Source**

---

✔ Local First &nbsp;✔ Provider-Agnostic &nbsp;✔ NPU / GPU / CPU  
✔ Agent Skills &nbsp;✔ Tools & Hooks &nbsp;✔ LSP & MCP

---

## What is OpenForge?

OpenForge is an **AI coding agent** that runs entirely on your machine. It reads, writes, edits, and executes code through a tool-calling agent loop powered by local LLMs. No cloud, no API keys, no telemetry.

It works with **any local inference runtime** — OpenVINO for Intel NPU/GPU, Ollama for GGUF models, llama.cpp for CPU-first inference, vLLM for multi-GPU setups, or LM Studio for a GUI experience.

## Why OpenForge?

| Problem | Solution |
|---------|----------|
| AI coding tools require cloud APIs | **100% local inference** — OpenVINO, Ollama, llama.cpp |
| Agent tools are hard-coded per provider | **Provider-agnostic** agent loop — swap runtime, keep tools |
| No control over what the agent does | **Hooks system** — PreToolUse/PostToolUse shell scripts |
| LSP integration is IDE-specific | **Built-in LSP Manager** — gopls, ts, rust, python auto-detect |
| Prompt engineering is repetitive | **Reusable Skills** — SKILL.md + YAML pipelines |
| Proprietary MCP tools are vendor-locked | **Native MCP client** — open protocol, any server |

## Quick Start

```bash
# Install a local runtime
openforge provider install ollama
ollama pull llama3.2:3b

# Start the agent
openforge

# Or start the API server
openforge serve
```

## Architecture

OpenForge is built in **Go** with a layered, modular architecture:

```
Agent Loop → Tool System → Provider Manager → Local Runtimes
                ↓
        Hooks · LSP · MCP · Skills
```

### Layers

| Layer | Package | Responsibility |
|-------|---------|---------------|
| **Agent Loop** | `internal/agent/` | input → LLM → tool calls → execute → feedback |
| **Tool System** | `internal/tool/` | bash, view, write, edit, grep, glob, ls, todos, fetch |
| **Provider Manager** | `internal/pm/` | Auto-discovery, chain selection, lifecycle (5 providers) |
| **Hooks** | `internal/hooks/` | Pre/post tool-use shell scripts |
| **LSP** | `internal/lsp/` | Auto-detect + manage language servers |
| **MCP** | `internal/mcp/` | Model Context Protocol client |
| **Skills** | `internal/skill/` | YAML pipelines + SKILL.md loader |
| **TUI** | `internal/tui/` | Bubble Tea terminal UI with streaming |
| **Server** | `internal/server/` | Gin HTTP server, OpenAI-compatible API |
| **Config** | `internal/config/` | Viper YAML config with env override |

## Session de 5 Minutos

1. **[Instale](docs/getting-started/provider-installation.md)** um runtime local
2. **[Inicie](docs/getting-started/quickstart.md)** o agente ou servidor
3. **[Conecte](docs/community/dashboard.md)** qualquer cliente OpenAI-compatível
4. **[Estenda](docs/skills/creating-skills.md)** com Skills e Hooks
5. **[Contribua](docs/contributing.md)** com código, docs ou feedback

## Ecossistema

- **OpenCode** — provedor nativo (via API OpenAI-compatível)
- **VS Code** — extensão autocomplete
- **Superpowers** — skills spec-driven
- **Qualquer cliente OpenAI** — drop-in replacement

## Licença

Apache 2.0 — Free for personal and commercial use.

---

[Guia de Providers](docs/getting-started/provider-installation.md) · [Arquitetura](docs/architecture.md) · [ADRs](docs/adr/) · [API](docs/api-reference.md) · [Contribuindo](docs/contributing.md)
