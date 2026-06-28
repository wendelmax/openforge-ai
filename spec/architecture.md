# Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────┐
│                     OpenCode / Superpowers               │
│                     (IDE, CLI, Editor)                   │
└──────────────────────┬──────────────────────────────────┘
                       │ HTTP / IPC
                       ▼
┌─────────────────────────────────────────────────────────┐
│                  OpenForge API Layer                     │
│                  (Gin HTTP Server)                       │
├─────────────────────────────────────────────────────────┤
│  /chat  /completion  /embedding  /rerank /model/...     │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│                    AI Engine (Core)                      │
├─────────────────────────────────────────────────────────┤
│  Session Manager     │    Pipeline Manager               │
│  Context Manager     │    Skill Executor                 │
│  Cache Manager       │    Telemetry                      │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│                     Provider Layer                       │
├─────────────────────────────────────────────────────────┤
│  OpenVINO Provider     │    Embedding Provider           │
│  Rerank Provider       │    (future: Plugin API)         │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│                  OpenVINO Runtime                        │
├─────────────────────────────────────────────────────────┤
│  Core    │  Inference  │   Model Loading    │  Devices   │
└──────────────────────┬──────────────────────────────────┘
                       │
         ┌─────────────┼─────────────┐
         ▼             ▼             ▼
       CPU           GPU           NPU
    (x86_64)    (Intel Iris)   (Intel AI Boost)
```

## Layer Responsibilities

### API Layer
- Expor endpoints REST para comunicação com OpenCode/Superpowers
- Gerenciar autenticação e rate limiting (quando aplicável)
- Validar requisições e serializar respostas
- Documentação OpenAPI 3.0 automática

### AI Engine
- Orquestrar fluxos de inferência (chat, completion, embedding, rerank)
- Gerenciar sessões de conversa e contexto
- Executar skills e pipelines multi-etapa
- Gerenciar cache de embeddings e respostas
- Coletar métricas e traces para telemetria

### Provider Layer
- Abstrair o OpenVINO Runtime para o Engine
- Gerenciar ciclo de vida de modelos (load, unload, reload)
- Detectar e selecionar dispositivos otimamente
- Interface de plugin para providers futuros (futuro)

### OpenVINO Runtime
- Carregar e executar modelos OpenVINO IR (.xml + .bin)
- Gerenciar dispositivos (CPU, GPU, NPU)
- Realizar inferência síncrona e assíncrona
- Gerenciar memória e recursos do hardware

## Communication Patterns

| Camadas | Protocolo | Formato |
|---------|-----------|---------|
| OpenCode ↔ API | HTTP/1.1 | JSON |
| API ↔ Engine | Chamada de função interna | structs Go |
| Engine ↔ Provider | Interface Go | structs Go |
| Provider ↔ OpenVINO | CGO / OpenVINO C API | buffers |

## Dependency Rules

```
API Layer → Engine → Provider → OpenVINO Runtime
     ↓           ↓         ↓
  NUNCA depende  NUNCA depende  NUNCA depende
  de camadas     de API         de API ou Engine
  superiores     Layer          Layer
```

## Cross-Cutting Concerns

- **Logging**: logs estruturados via slog (structured logging)
- **Config**: carregamento de config por camada com merge hierárquico
- **Errors**: erros tipados com wrapped context
- **Metrics**: contadores e histogramas via OpenTelemetry
- **Tracing**: spans distribuídos para operações críticas
