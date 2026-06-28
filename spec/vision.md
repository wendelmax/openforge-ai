# Product Vision

**OpenForge** é um framework open source para desenvolvimento assistido por IA utilizando exclusivamente o OpenVINO Runtime como mecanismo de inferência.

## Propósito

Capacitar desenvolvedores com ferramentas de IA generativa que rodam 100% localmente, eliminando dependência de serviços externos, APIs de nuvem ou soluções proprietárias.

## Público-Alvo

- Desenvolvedores que necessitam de assistência offline
- Equipes que trabalham em ambientes air-gapped
- Projetos que exigem privacidade total dos dados
- Engenheiros que precisam de baixa latência em inferência local
- Comunidade OpenCode e Superpowers

## Diferenciais

- **Zero dependências externas**: não requer Ollama, LM Studio, APIs cloud ou qualquer serviço remoto
- **Hardware Intel completo**: suporte a CPU, GPU integrada e NPU (Neural Processing Unit)
- **Plug-and-play**: detecção automática do melhor dispositivo disponível
- **Modular por design**: sistema de plugins e skills reutilizáveis
- **Performance**: cache inteligente, inferência assíncrona, otimizações OpenVINO
- **Compatibilidade total**: funciona como provider nativo para OpenCode e Superpowers

## Objetivos Estratégicos

1. Execução totalmente offline sem degradação de experiência
2. Arquitetura modular e extensível baseada em plugins
3. Alta performance com inferência otimizada para hardware Intel
4. Compatibilidade total com o ecossistema OpenCode e Superpowers
5. Cobertura de testes mínima de 90%
6. Documentação completa em português e inglês
7. Distribuição via Docker, GitHub Releases e package managers

## Não-Escopo

- Não substitui soluções cloud para fine-tuning ou treinamento
- Não oferece interface gráfica própria (foco em CLI, API e integração com IDEs)
- Não implementa funcionalidades de MLOps ou gerenciamento de datasets
