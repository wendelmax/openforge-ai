# ADR-0001: OpenVINO como único mecanismo de inferência

## Status

Accepted

## Context

O projeto precisa de um runtime de inferência local para execução de modelos de linguagem. As opções consideradas foram:

- **Ollama**: dependência externa, não oferece controle fino sobre dispositivos (CPU/GPU/NPU), overhead de processo separado
- **LM Studio**: interface gráfica, não headless, dependência de UI
- **llama.cpp**: excelente performance, mas sem suporte nativo a NPU Intel
- **ONNX Runtime**: suporte amplo, mas sem otimizações específicas para hardware Intel
- **OpenVINO**: runtime oficial da Intel com suporte a CPU, GPU integrada, GPU discreta e NPU

## Decisão

Utilizar exclusivamente o **OpenVINO Runtime** como mecanismo de inferência.

## Justificativa

1. **Hardware Intel completo**: OpenVINO é o único runtime que suporta nativamente CPU, GPU (Iris, Arc) e NPU (AI Boost) Intel
2. **Otimizações específicas**: OpenVINO inclui otimizações por hardware (throughput, latência, consumo)
3. **C/C++ API**: permite bindings diretos com Go via CGO sem camadas extras
4. **Modelo de deployment**: modelo IR (Intermediate Representation) é aberto e portável
5. **Zero dependências externas**: não requer Ollama, Docker ou serviços cloud
6. **Foco no hardware alvo**: Intel é o hardware mais comum em estações de trabalho de desenvolvimento

## Consequências

### Positivas
- Suporte completo ao ecossistema Intel (CPU, GPU, NPU)
- Performance otimizada sem configuração manual
- Instalação simplificada (runtime único)
- Documentação oficial extensa

### Negativas
- Vendor lock-in com Intel (usuários AMD/Apple têm suporte limitado a CPU)
- Complexidade de CGO (bindings entre Go e C++)
- Modelos precisam ser convertidos para formato IR (OpenVINO)
- Curva de aprendizado da API OpenVINO

## Alternativas Rejeitadas

- Ollama: rejeitado por depender de processo externo e não oferecer controle direto sobre dispositivos
- llama.cpp: rejeitado por falta de suporte a NPU Intel, que é um diferencial chave do projeto
- ONNX Runtime: rejeitado por falta de otimizações específicas para Intel NPU
