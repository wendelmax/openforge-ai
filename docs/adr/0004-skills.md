# ADR-0004: Sistema de Skills

## Status

Proposed

## Context

O OpenForge precisa de um mecanismo para execução de tarefas compostas por múltiplos passos de IA. Por exemplo: "resumir documento" pode ser um skill que faz embedding do documento, passa por um prompt de sumarização, e retorna o resultado formatado.

Skills são diferentes de plugins:
- **Plugins**: extendem a capacidade do provider (novos backends, novos dispositivos)
- **Skills**: pipelines de IA reutilizáveis que orquestram chamadas ao engine

## Decisão

Implementar skills como arquivos de configuração (YAML) que definem um pipeline de passos, executados sequencialmente pelo Skill Executor.

## Justificativa

1. **Configuração sobre código**: usuários podem criar skills sem escrever Go
2. **Versionável**: skills em YAML podem ser versionadas em git
3. **Compartilhável**: skills podem ser distribuídas como arquivos
4. **Simplicidade**: pipeline linear é mais fácil de entender e depurar
5. **Extensível**: passos podem ser adicionados no futuro (condicionais, loops)

## Consequências

### Positivas
- Usuários podem criar skills customizadas sem compilar código
- Skills são portáveis entre instalações
- Fácil compartilhamento via repositórios de skills
- Pipeline visualmente claro no YAML

### Negativas
- YAML pode se tornar complexo para pipelines avançados
- Sem validação em tempo de compilação
- Performance limitada por ser interpretado (não compilado)
- Necessidade de um schema de validação para YAML

## Exemplo de Skill

```yaml
name: summarize
description: Summarize a document
steps:
  - type: prompt
    system: "Summarize the following text concisely:"
    user: "{{input}}"
    model: llama-3.2-3b
    output: summary

  - type: format
    template: |
      ## Summary
      {{summary}}
    output: result
```
