# Coding Conventions

## General

- Toda classe pública deve possuir interface
- Toda dependência deve ser injetada (DI)
- Nunca usar Singleton como padrão de design
- Nunca usar variáveis globais
- Sempre usar async/goroutines para operações blocking
- Logs estruturados via `log/slog`
- Erros tipados com `fmt.Errorf("context: %w", err)`

## Go-Specific

- Usar `go 1.23+` com toolchain atual
- Nomes de pacotes em minúsculo, sem underscore
- Arquivos nomeados no singular (`model.go`, não `models.go`)
- Testes usando `testing` padrão + `testify/assert`
- Benchmarks obrigatórios para funções críticas
- Comentários GoDoc em toda exportação
- Evitar `init()` — preferir inicialização explícita
- Preferir `context.Context` como primeiro parâmetro
- Retornar `(result, error)` — nunca panics

## Architecture

- Seguir Clean Architecture (camadas: handlers → usecases → providers)
- Seguir SOLID principles rigorosamente
- Interfaces pequenas (1-3 métodos) no pacote consumidor
- Dependências apontam para abstrações, não implementações
- Pacotes sem dependência circular (ferramenta `go vet`)

## Testing

- Testes unitários obrigatórios para toda função exportada
- Testes de integração para camada de provider
- Mocks gerados via `mockgen` (ou manual quando simples)
- Nomenclatura: `TestFuncName_Scenario_ExpectedBehavior`
- Fixtures em arquivos `testdata/`
- Cobertura mínima: 90%

## Code Organization

```
pkg/      → bibliotecas reutilizáveis (públicas)
internal/ → implementação interna (não exportada)
cmd/      → entrypoints (main)
hack/     → scripts e ferramentas de desenvolvimento
```

## Documentation

- GoDoc em toda função, tipo e constante exportada
- README.md na raiz e em subpacotes quando necessário
- ADRs em `docs/adr/` para decisões arquiteturais
- Swagger/OpenAPI documentado no código

## Git

- Commits semânticos: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
- Branches: `feature/`, `bugfix/`, `release/`
- PRs com descrição clara do que foi alterado e por quê
- Nunca commitar secrets, tokens ou paths absolutos

## Configuration

- Config via YAML + variáveis de ambiente (override)
- Hierarchy: defaults → config.yaml → env vars → CLI flags
- Sensitive config apenas via env vars
- Toda config validada na inicialização

## API

- Versionamento via path prefix (`/v1/`, `/v2/`)
- Compatibilidade retroativa obrigatória
- Breaking changes apenas em major version
- Erros sempre no formato padrão
- Headers de request ID para tracing
