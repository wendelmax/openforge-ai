# Roadmap

## Sprint 1 — Runtime Foundation
- Setup do projeto Go (module, estrutura de pastas)
- Implementação da interface de runtime
- OpenVINO CGO bindings básicos
- Carregamento e descarregamento de modelos
- Testes unitários iniciais

## Sprint 2 — Device Detection
- Detecção de CPU
- Detecção de GPU Intel
- Detecção de NPU Intel
- Algoritmo de seleção automática de dispositivo
- Fallback manual de dispositivo
- Testes de integração com hardware

## Sprint 3 — Provider Layer
- Implementação do OpenVINO provider
- Abstração de provider (interface)
- Inference síncrona
- Inference assíncrona
- Cache de modelo em memória

## Sprint 4 — API Server
- Setup do Gin HTTP server
- Endpoint /v1/chat
- Endpoint /v1/completion
- Endpoint /v1/embeddings
- Endpoint /v1/models
- Endpoint /v1/devices
- Endpoint /v1/health
- Suporte a SSE streaming
- Configuração via YAML

## Sprint 5 — Cache & Session
- Cache de embeddings via SQLite
- Cache de respostas via SQLite
- Gerenciamento de sessões de conversa
- Gerenciamento de contexto
- Invalidação de cache

## Sprint 6 — Embeddings & Rerank
- Pipeline de embeddings
- Modelo BGE Small integrado
- Endpoint /v1/rerank
- Busca semântica básica
- Testes de qualidade de embedding

## Sprint 7 — Skills
- Skill definition format (YAML/JSON)
- Skill loader
- Skill executor (pipeline engine)
- Skills reutilizáveis padrão
- Endpoint para execução de skills

## Sprint 8 — Plugin System
- Plugin interface definition
- Plugin loader (so/dll)
- Plugin lifecycle management
- Provider plugins
- Example plugin

## Sprint 9 — Observability
- OpenTelemetry integration
- Métricas de performance
- Logs estruturados detalhados
- Profiling endpoints (pprof)
- Health check avançado

## Sprint 10 — Benchmark & Release
- Benchmark de modelos
- Relatório de benchmark exportável
- GoReleaser configuration
- Docker multi-stage build
- Documentação completa
- Release v1.0.0-alpha

## Future (Post v1.0)
- VS Code extension integration
- RAG pipeline completo
- Multi-model serving
- GPU acceleration tuning
- Web UI básica (admin/monitor)
- Windows native packaging
- Homebrew/Linuxbrew formula
