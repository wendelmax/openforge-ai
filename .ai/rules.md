# AI Agent Rules

## Golden Rules

1. **Nunca criar código duplicado** — sempre abstrair em interfaces/funções reutilizáveis
2. **Sempre reutilizar interfaces existentes** — verificar se já existe interface antes de criar nova
3. **Sempre escrever testes** — unit tests obrigatórios, integration tests para providers
4. **Sempre documentar APIs** — GoDoc em toda exportação, OpenAPI em todo endpoint
5. **Nunca quebrar compatibilidade** — mudanças devem ser backward compatible
6. **Sempre seguir Clean Architecture** — camadas bem definidas, dependências apontando para dentro
7. **Sempre usar SOLID** — Single Responsibility, Open-Closed, Liskov, Interface Segregation, Dependency Inversion
8. **Sempre executar testes antes de considerar uma tarefa completa** — `go test ./... && go vet ./...`
9. **Sempre ler os ADRs antes de tomar decisões arquiteturais** — contexto é obrigatório
10. **Nunca assumir hardware específico** — sempre tratar device unavailable graciosamente

## Code Generation Rules

1. **Interfaces primeiro** — definir interface antes da implementação
2. **Tipos fortes** — evitar `interface{}`, preferir tipos concretos ou genéricos (Go 1.23)
3. **Errors são valores** — nunca ignorar erros, sempre propagar com contexto
4. **Context primeiro** — toda função blocking deve receber `context.Context` como primeiro parâmetro
5. **Imports organizados** — stdlib → third-party → internal
6. **Nomes descritivos** — evitar abreviações (exceto convenções como `ctx`, `wg`, `err`)
7. **Zero values úteis** — structs devem ter zero value utilizável quando possível
8. **Concorrência explícita** — usar goroutines com sincronização clara (WaitGroup, mutex, channels)

## File Creation Rules

1. Um arquivo por tipo principal (ex: `model.go` contém `Model`, `model_test.go` contém testes)
2. Testes no mesmo pacote (white-box testing)
3. Testes de integração com build tag `//go:build integration`
4. Arquivos de mock em `internal/mocks/`
5. Configurações de exemplo em `config/example.yaml`

## Decision Making

1. Ao encontrar ambiguidade na spec, consultar ADRs primeiro
2. Se ADR não cobre o caso, criar novo ADR antes de implementar
3. Preferir soluções mais simples (YAGNI)
4. Não adicionar dependências externas sem necessidade justificada
5. Performance measurements devem ser baseadas em benchmark, não suposição

## Communication

1. Commits semânticos obrigatórios
2. Mensagens de commit em português ou inglês (consistente no repositório)
3. Comentários explicam "por quê", não "o quê"
4. Documentação em português e inglês quando possível
