# ADR-0003: Arquitetura baseada em plugins

## Status

Proposed

## Context

O projeto precisa ser extensível para suportar:
- Diferentes provedores de inferência (futuramente)
- Skills personalizadas pela comunidade
- Integrações com IDEs e ferramentas
- Middleware e hooks personalizados

Arquiteturas consideradas:
- **Monolítica**: tudo no mesmo processo, sem plugins
- **Plugin nativo (Go plugin)**: pacote `plugin` do Go (shared objects)
- **Plugin via WASM**: WebAssembly para isolamento e segurança
- **Plugin via subprocesso**: comunicação via IPC/RPC

## Decisão

Implementar um sistema de plugins baseado em interfaces Go, com carregamento via shared objects (`.so`/`.dll`) usando o pacote `plugin` do Go para fase 1, e WASM como plano futuro.

## Justificativa

1. **Simplicidade**: Go plugin não requer infraestrutura adicional
2. **Performance**: chamada direta sem serialização ou IPC
3. **Mesma linguagem**: plugins escritos em Go, mesma toolchain
4. **Isolamento parcial**: cada plugin em seu próprio shared object
5. **Familiaridade**: padrão comum em projetos Go (HashiCorp, etc.)

## Consequências

### Positivas
- Extensível sem modificar o core
- Comunidade pode criar plugins independentemente
- Interface bem definida para providers
- Possibilidade de plugins externos (empresas/parceiros)

### Negativas
- Go plugin é limitado (mesma versão de Go, Linux/macOS apenas, sem Windows nativo)
- Sem isolamento de segurança (plugin roda no mesmo processo)
- Complexidade de versionamento entre core e plugins
- Debugging mais difícil com plugins carregados

## Alternativas Rejeitadas

- **Monolítica**: rejeitada por falta de extensibilidade
- **Subprocesso**: rejeitado por latência de IPC e complexidade de deployment
- **WASM**: rejeitado para fase 1 por maturidade do ecossistema Go-WASM; considerado para fase 2

## Notas

- Fase 1: plugins .so/.dll com interface Go
- Fase 2 (futuro): suporte a WASM plugins para isolamento e segurança
- A interface de plugin será versionada semver
