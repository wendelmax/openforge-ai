# OpenForge

**AI Runtime for Developers — 100% Local, 100% OpenVINO**

---

✔ Local First &nbsp;✔ OpenVINO Native &nbsp;✔ CPU / GPU / NPU  
✔ Skills &nbsp;✔ Spec Driven &nbsp;✔ Open Source

---

## What is OpenForge?

OpenForge is an open-source AI inference framework that runs **exclusively on OpenVINO Runtime**. It lets you run LLMs, embeddings, and reranking models **entirely offline** on Intel hardware — no cloud, no Ollama, no LM Studio.

## Why OpenForge?

| Problem | Solution |
|---------|----------|
| Running LLMs locally requires complex setup | One binary, zero dependencies |
| Ollama/LM Studio add overhead | Direct OpenVINO integration |
| GPU/NPU detection is manual | Automatic device selection |
| Prompt engineering is repetitive | Reusable Skills |
| AI-generated code lacks quality | Spec-Driven Development |

## Quick Start

```bash
# Install
curl -fsSL https://openforge.ai/install.sh | bash

# Start server
openforge serve

# Chat
curl http://localhost:9090/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"model":"llama-3.2-3b","messages":[{"role":"user","content":"Hello"}]}'
```

## 5-Minute Tour

1. **Install** the binary
2. **Download** an OpenVINO model
3. **Start** the server
4. **Connect** OpenCode or any OpenAI-compatible client
5. **Extend** with Skills

## Performance

| Model | CPU | GPU | NPU |
|-------|----:|----:|----:|
| Phi-3 Mini | 18 t/s | 32 t/s | 41 t/s |
| Llama 3.2 3B | 12 t/s | 22 t/s | 28 t/s |
| BGE Small (embed) | 15ms | 8ms | 5ms |

## Ecosystem

- **OpenCode** — native provider
- **Superpowers** — spec-driven agent
- **VS Code** — autocomplete extension
- **OpenAI API** — drop-in replacement

## License

Apache 2.0 — Free for personal and commercial use.
