# Guia de Instalação — OpenForge Agent

> Rode agentes de IA 100% locais. NPU, GPU, CPU — zero cloud.

## Pré-requisitos

| SO | Versão | Hardware recomendado |
|----|--------|---------------------|
| Windows 11 | 24H2+ | Intel Core Ultra (NPU) / GPU NVIDIA / CPU moderna |
| Ubuntu / Debian | 22.04+ | Intel Core Ultra / GPU NVIDIA / CPU moderna |
| WSL 2 | Ubuntu 22.04+ | Hardware da máquina host |
| macOS | 14+ | Apple Silicon ou Intel |

**Mínimo**: 8GB RAM, 20GB disco.  
**Recomendado**: 16GB+ RAM, SSD.

---

## Instalação Rápida (Auto-Detect)

```bash
# Baixa e instala o OpenForge
curl -fsSL https://openforge.ai/install.sh | bash

# Auto-detect: encontra hardware, instala runtime, baixa modelo
openforge init --auto

# Pronto
openforge
```

O `openforge init --auto` faz:
1. Detecta hardware (CPU, GPU, NPU)
2. Detecta runtimes já instalados
3. Instala o melhor runtime disponível
4. Baixa modelo recomendado (~3GB)
5. Gera configuração otimizada
6. Testa inferência

---

## Instalação por Ambiente

### Opção A — Windows 11 Nativo (NPU Intel)

Roteiro para Intel Core Ultra com NPU (AI Boost):

```powershell
# 1. Instalar OpenVINO Runtime
pip install openvino openvino-genai

# 2. Instalar NPU drivers (requer Core Ultra)
# Download: https://www.intel.com/content/www/us/en/download/794734/intel-npu-driver-windows.html

# 3. Converter modelo para OpenVINO IR
pip install optimum[openvino]
optimum-cli export openvino --model microsoft/Phi-3.5-mini-instruct ./models/phi-3.5-mini

# 4. Configurar OpenForge
openforge config init --provider openvino --model-path ./models --device NPU

# 5. Testar
openforge --model phi-3.5-mini "Explique computação quântica em uma frase"
```

**Performance esperada**: 35-45 tok/s (NPU) — sem GPU dedicada.

### Opção A.2 — Windows 11 com GPU NVIDIA

```powershell
# 1. Instalar Ollama
winget install Ollama.Ollama
# Reinicie o terminal

# 2. Iniciar Ollama
ollama serve

# Em outro terminal:
# 3. Baixar modelo
ollama pull llama3.2:3b

# 4. Configurar OpenForge
openforge config init --provider ollama

# 5. Testar
openforge "Crie uma função em Go que reverte uma string"
```

### Opção A.3 — Windows 11 sem GPU (CPU only)

```powershell
# 1. Instalar Ollama (usa CPU automaticamente)
winget install Ollama.Ollama

# 2. Baixar modelo leve (funciona bem em CPU)
ollama pull phi3:mini

# 3. OpenForge
openforge init --auto
```

### Opção B — WSL 2 (Ubuntu)

```bash
# 1. Instalar Ollama no WSL
curl -fsSL https://ollama.com/install.sh | sh

# 2. Iniciar Ollama
ollama serve

# Em outra sessão:
# 3. Baixar modelo
ollama pull llama3.2:3b

# 4. Instalar OpenForge
curl -fsSL https://openforge.ai/install.sh | bash

# 5. Init
openforge init --provider ollama

# 6. Rodar
openforge
```

**Nota WSL**: GPU funciona via Ollama com drivers NVIDIA WSL (`nvidia-smi` deve funcionar dentro do WSL).

### Opção C — Linux Nativo

```bash
# ---- Rota 1: Ollama (mais fácil) ----

# 1. Instalar Ollama
curl -fsSL https://ollama.com/install.sh | sh

# 2. Baixar modelo
ollama pull llama3.2:3b

# 3. OpenForge
openforge init --provider ollama

# ---- Rota 2: OpenVINO (máxima performance em Intel) ----

# 1. Instalar OpenVINO
pip install openvino openvino-genai

# 2. Intel GPU driver
sudo apt install intel-opencl-icd

# 3. Converter modelo
pip install optimum[openvino]
optimum-cli export openvino --model microsoft/Phi-3.5-mini-instruct ./models/phi-3.5-mini

# 4. OpenForge
openforge init --provider openvino --model-path ./models --device GPU

# ---- Rota 3: vLLM (máxima performance em NVIDIA) ----

# 1. Instalar vLLM
pip install vllm

# 2. Iniciar servidor
vllm serve meta-llama/Llama-3.2-3B-Instruct --port 8000

# 3. OpenForge
openforge init --provider openai-compat --endpoint http://localhost:8000/v1
```

---

## Comandos de Gestão

```bash
# Ver todos os providers detectados
openforge provider list

# Detalhes de um provider específico
openforge provider info ollama

# Health check
openforge provider status

# Instalar um provider
openforge provider install ollama

# Guia passo a passo interativo
openforge provider guide

# Auto-detecção completa
openforge provider detect
```

---

## Configuração Manual

Arquivo `~/.openforge/config.yaml` ou `./openforge.yaml`:

```yaml
providers:
  # Ordem de tentativa
  chain: [openvino, ollama, llamacpp, vllm, lmstudio]

  # Provider padrão por tipo de tarefa
  workloads:
    chat: auto      # usa o melhor disponível
    embed: auto
    rerank: auto
    code: auto

  # --- OpenVINO ---
  openvino:
    enabled: true
    model_path: ./models
    device: auto    # auto = NPU > GPU > CPU

  # --- Ollama ---
  ollama:
    enabled: true
    endpoint: http://localhost:11434
    auto_pull: true  # baixa modelos automaticamente

  # --- llama.cpp ---
  llamacpp:
    enabled: true
    endpoint: http://localhost:8080/v1

  # --- vLLM ---
  vllm:
    enabled: true
    endpoint: http://localhost:8000/v1

  # --- LM Studio ---
  lmstudio:
    enabled: true
    endpoint: http://localhost:1234/v1
```

---

## Modelos Recomendados por Hardware

| Hardware | Modelo | Tamanho | Tok/s esperado | Provider |
|----------|--------|---------|----------------|----------|
| NPU Intel | Phi-3.5 Mini | 3.8B | 35-45 | OpenVINO |
| NPU Intel | Llama 3.2 3B | 3.2B | 28-35 | OpenVINO |
| GPU Intel Arc | Llama 3.2 3B | 3.2B | 30-40 | OpenVINO |
| GPU NVIDIA 8GB+ | Llama 3.1 8B | 8B | 50-80 | Ollama/vLLM |
| GPU NVIDIA 6GB | Qwen2.5 7B | 7B | 30-50 | Ollama |
| GPU NVIDIA 4GB | Llama 3.2 3B | 3.2B | 40-60 | Ollama |
| CPU moderna (16GB) | Llama 3.2 3B | 3.2B | 10-18 | Ollama/llama.cpp |
| CPU moderna (8GB) | Phi-3 Mini | 3.8B | 8-12 | Ollama |
| CPU antiga (8GB) | TinyLlama 1.1B | 1.1B | 15-25 | Ollama |
| Apple M1/M2 16GB | Llama 3.2 3B | 3.2B | 25-35 | Ollama |

---

## Troubleshooting

### Ollama não conecta
```bash
# Verifique se está rodando
curl http://localhost:11434/api/tags

# Inicie se necessário
ollama serve
```

### OpenVINO não encontra NPU
```powershell
# Windows: verifique driver NPU
# Gerenciador de Dispositivos > Intel AI Boost > deve aparecer sem warning

# Reinstale drivers NPU
# https://www.intel.com/content/www/us/en/download/794734
```

### "no provider available"
```bash
# Diagnóstico completo
openforge provider detect

# Instale pelo menos um runtime
openforge provider install ollama
```

### WSL: GPU não funciona
```bash
# Instale driver NVIDIA para WSL no Windows host
# Depois verifique no WSL:
nvidia-smi
# Deve mostrar a GPU
```

### Modelo muito lento
- Use modelo menor (3B em vez de 8B)
- Verifique se está usando GPU: `ollama ps` (mostra o device)
- OpenVINO: tente `--device GPU` em vez de `--device NPU`
- Considere Q4_K_M quantization (modelos GGUF)

---

## Próximos Passos

```bash
# Ler documentação completa
openforge docs

# Criar skills personalizadas
openforge skill init

# Ver benchmark de performance
openforge benchmark

# Rodar servidor HTTP (OpenAI-compatível)
openforge serve

# Configurar OpenCode/VS Code para usar OpenForge
openforge config opencode
```

---

💘 **OpenForge — Agente de IA 100% local para desenvolvedores.**
