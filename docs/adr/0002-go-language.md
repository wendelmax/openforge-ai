# ADR-0002: Go como linguagem principal

## Status

Accepted

## Context

O projeto requer uma linguagem que combine performance, segurança de tipos, boa experiência de desenvolvimento e facilidade para integração com C/C++ (para OpenVINO).

Opções consideradas:

- **Python**: excelente para IA, mas problemas de performance, tipagem fraca, GIL, deployment complexo
- **Rust**: performance máxima, mas curva de aprendizado alta, complexidade de CGO alternativa, ecossistema imaturo para HTTP/API
- **C++**: performance máxima, mas verbosidade, segurança de memória manual, build complexo
- **Go**: boa performance, goroutines para concorrência, CGO para OpenVINO, tipagem forte, build simples (binário único)

## Decisão

Utilizar **Go** como linguagem principal do projeto.

## Justificativa

1. **CGO**: integração direta com OpenVINO C API sem camadas adicionais
2. **Concorrência**: goroutines são ideais para inferência assíncrona e streaming
3. **Build estático**: binário único sem dependências de runtime
4. **Performance**: compilado, tipado, com boa performance para API serving
5. **Ecossistema**: Gin para HTTP, SQLite, testify, ferramentas maduras
6. **Produtividade**: build rápido, testes integrados, formatação padrão
7. **Cross-compilação**: fácil build para Linux, Windows, macOS

## Consequências

### Positivas
- Binário único e portável
- Concorrência nativa eficiente
- CGO permite bindings diretos com OpenVINO
- Tipagem forte reduz erros em runtime
- Excelente ferramental (linter, tests, profiling)

### Negativas
- CGO adiciona complexidade ao build (cross-compilação limitada)
- Garbage collector pode impactar latência
- Ecossistema de IA/ML em Go é menor que Python
- Menos bibliotecas para processamento de linguagem natural

## Alternativas Rejeitadas

- Python: rejeitado por problemas de performance, GIL e deployment complexo
- Rust: rejeitado por complexidade de FFI com OpenVINO e curva de aprendizado
- C++: rejeitado por verbosidade e falta de segurança de memória
