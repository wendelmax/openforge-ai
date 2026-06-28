# ADR-0006: Documentação de API com OpenAPI 3.0

## Status

Accepted

## Context

O OpenForge expõe uma API REST que será consumida por:
- OpenCode
- Superpowers
- VS Code extension
- Scripts e ferramentas de usuários
- Outros agentes

É necessário documentar esta API de forma padronizada e acessível.

## Decisão

Documentar a API usando OpenAPI 3.0 com geração automática a partir do código Go, utilizando a biblioteca `swaggo/swag`.

## Justificativa

1. **Padrão da indústria**: OpenAPI é o formato mais adotado para documentação de APIs REST
2. **Geração automática**: `swaggo/swag` gera spec OpenAPI a partir de comentários Go
3. **UI interativa**: Swagger UI pode ser servida pelo próprio servidor
4. **Code generation**: clientes podem ser gerados a partir da spec
5. **Versionamento**: spec pode ser versionada junto com o código

## Consequências

### Positivas
- Documentação sempre sincronizada com o código
- UI interativa para testes durante desenvolvimento
- Geração de clientes para outras linguagens
- Facilita integração com OpenCode/Superpowers

### Negativas
- Comentários adicionais necessários no código Go
- Build dependency (swag CLI)
- Pode gerar specs muito verbosas
- swaggo tem limitações com tipos complexos

## Alternativas Rejeitadas

- **Manuscrito**: rejeitado por ficar dessincronizado do código
- **API Blueprint**: rejeitado por menor adoção e ferramentas menos maduras
- **GraphQL**: rejeitado por overhead desnecessário para API de inferência
