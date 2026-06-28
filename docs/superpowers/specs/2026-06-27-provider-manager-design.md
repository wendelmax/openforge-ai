# Provider Manager — Local LLM Runtime Management

> Especificação para auto-descoberta, instalação e gestão de runtimes de inferência local.

## Filosofia

O OpenForge é **provider-agnostic** por design. O usuário não deve precisar saber qual runtime está por baixo — ele só quer que o modelo rode no hardware disponível com a melhor performance possível.

```
Zero-config ideal:
  Detectar hardware → Detectar runtimes instalados → Escolher o melhor → Rodar

Fallback automático:
  OpenVINO (NPU) → OpenVINO (GPU) → Ollama (GPU/CPU) → llama.cpp (CPU) → Erro amigável
```

## Providers Suportados

| Provider | Runtime | Modelos | Hardware | Interface |
|----------|---------|---------|----------|-----------|
| **OpenVINO** | OpenVINO Runtime C++ | `.xml` + `.bin` (IR) | CPU, GPU Intel, **NPU Intel** | CGO nativo |
| **Ollama** | llama.cpp via Ollama | GGUF | CPU, GPU (CUDA/Vulkan) | HTTP REST |
| **llama.cpp** | llama.cpp server | GGUF | CPU, GPU (CUDA/Metal/Vulkan) | HTTP REST |
| **vLLM** | vLLM Python | HuggingFace, safetensors | GPU (CUDA) | OpenAI-compat HTTP |
| **LM Studio** | LM Studio (bundled) | GGUF | CPU, GPU (CUDA/Metal) | OpenAI-compat HTTP |

## Provider Chain Padrão

A ordem de resolução segue performance + disponibilidade esperada em hardware Intel:

```yaml
providers:
  chain: [openvino, ollama, llamacpp, vllm, lmstudio]
  default: auto  # "auto" = percorre a chain até achar um provider disponível
```

Para cada **workload**, a resolução é:

1. Se `request.provider` ou `request.device` for especificado → usa esse
2. Se `config.workloads.<workload>.provider` existir → tenta esse
3. Senão → percorre `providers.chain` em ordem até achar um provider que:
   - Esteja instalado
   - Tenha o modelo solicitado
   - Suporte o workload (chat, embed, rerank, code)

## Auto-Discovery

Na inicialização, o Provider Manager verifica cada runtime:

### OpenVINO
- **Detecção**: Tenta carregar `openvino_c.dll` / `libopenvino_c.so`
- **Hardware**: `ov::Core::get_available_devices()` → CPU, GPU.0, GPU.1, NPU
- **Stub mode**: Se não encontrar OpenVINO, compila sem CGO

### Ollama
- **Detecção**: `GET http://localhost:11434/api/tags` (timeout 2s)
- **Modelos**: Lista da API
- **Instalação**: `curl -fsSL https://ollama.com/install.sh | sh` ou `winget install Ollama`

### llama.cpp
- **Detecção**: `GET http://localhost:8080/v1/models` (timeout 2s)
- **Portas**: 8080, 8081, 8082 (scan)
- **Instalação**: Download do release ou `brew install llama.cpp`

### vLLM
- **Detecção**: `GET http://localhost:8000/v1/models` (timeout 2s)
- **Portas**: 8000, 8001
- **Instalação**: `pip install vllm`

### LM Studio
- **Detecção**: `GET http://localhost:1234/v1/models` (timeout 2s)
- **Windows**: `%LOCALAPPDATA%/LM Studio`
- **macOS**: `~/Applications/LM Studio.app`

## Provider Abstraction

```go
// Provider define a interface comum que todo runtime local implementa.
type Provider interface {
    // Info retorna metadados do provider.
    Info() ProviderInfo

    // Status retorna o estado atual do provider (disponível, ocupado, erro).
    Status(ctx context.Context) (*ProviderStatus, error)

    // Chat completa uma conversa síncrona.
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

    // ChatStream completa uma conversa com streaming.
    ChatStream(ctx context.Context, req *ChatRequest) (<-chan Token, error)

    // Embed gera embeddings para os textos de entrada.
    Embed(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error)

    // ListModels retorna os modelos disponíveis neste provider.
    ListModels(ctx context.Context) ([]Model, error)

    // LoadModel carrega um modelo (se o runtime suportar carga explícita).
    LoadModel(ctx context.Context, modelID string) error

    // UnloadModel descarrega um modelo da memória.
    UnloadModel(ctx context.Context, modelID string) error
}
```

## Auto-Instalação

O Provider Manager pode **guiar** (não forçar) a instalação de runtimes:

### Comando Interativo
```bash
openforge provider install         # Menu interativo: qual provider instalar?
openforge provider install ollama  # Instala Ollama direto
openforge provider detect          # Escaneia o sistema e mostra o que encontrou
```

### Fluxo de Instalação
1. Detecta SO e arquitetura
2. Mostra opções de runtime compatíveis com o hardware
3. Executa script de instalação apropriado
4. Verifica se instalou corretamente (health check)
5. Configura provider no `config.yaml`

### Scripts de Instalação Embutidos

Cada provider tem um script Go que executa a instalação padrão:

```go
// installers/ollama.go
func InstallOllama(ctx context.Context) error {
    switch runtime.GOOS {
    case "linux":
        return execCmd(ctx, "curl", "-fsSL", "https://ollama.com/install.sh", "|", "sh")
    case "darwin":
        return execCmd(ctx, "brew", "install", "ollama")
    case "windows":
        return execCmd(ctx, "winget", "install", "Ollama.Ollama")
    }
}
```

## Configuração

Exemplo completo de `config.yaml`:

```yaml
providers:
  # Chain de providers (ordem de tentativa)
  chain: [openvino, ollama, llamacpp, vllm, lmstudio]
  
  # Provider padrão por workload
  workloads:
    chat: auto
    code: auto
    embed: openvino     # embedding é leve, OpenVINO resolve bem
    rerank: openvino
  
  # Override por provider
  openvino:
    enabled: true
    model_path: ./models
    device: auto          # auto, CPU, GPU, NPU
    
  ollama:
    enabled: true
    endpoint: http://localhost:11434
    auto_pull: true       # baixa modelo automaticamente se não existir
    
  llamacpp:
    enabled: true
    endpoint: http://localhost:8080/v1
    
  vllm:
    enabled: true
    endpoint: http://localhost:8000/v1
    auto_start: true      # inicia vLLM se detectar GPU NVidia
    
  lmstudio:
    enabled: true
    endpoint: http://localhost:1234/v1
```

## Health Check & Startup

Na inicialização:

1. Escaneia `providers.chain` em ordem
2. Para cada provider, tenta health check (2s timeout)
3. Se falhar, marca como `unavailable` e continua
4. Marca primeiro provider disponível como `active`
5. Se nenhum estiver disponível, mostra erro amigável com instruções:
   ```
   ⚠ Nenhum runtime de IA local encontrado.
   
   Instale um dos seguintes:
   • Ollama:  openforge provider install ollama
   • OpenVINO: openforge provider install openvino
   • LM Studio: openforge provider install lmstudio
   
   Ou veja o guia completo: openforge provider guide
   ```

## CLI Commands

```bash
openforge provider                   # Lista providers com status
openforge provider list              # Lista detalhada
openforge provider install [name]    # Instala/guia instalação
openforge provider uninstall [name]  # Desinstala
openforge provider start [name]      # Inicia runtime (se aplicável)
openforge provider stop [name]       # Para runtime
openforge provider detect            # Auto-detecção de hardware + runtimes
openforge provider guide             # Guia completo de instalação
openforge config init                # Gera config.yaml com auto-detecção
```

## Device Priority & Smart Selection

Quando múltiplos providers oferecem o mesmo modelo, a escolha segue:

1. **Provider mais rápido para o workload** (benchmark histórico)
2. **Menor latência de startup** (OpenVINO > Ollama > vLLM)
3. **Hardware disponível** (NPU > GPU > CPU)
4. **Configuração explícita do usuário**

```
Exemplo: modelo "llama-3.2-3b" disponível em Ollama e OpenVINO

Workload "chat":
  1. OpenVINO GPU: 32 tok/s ✓ usa este
  2. Ollama GPU:   28 tok/s
  3. OpenVINO CPU: 18 tok/s

Workload "embed":
  1. OpenVINO NPU:  5ms ✓ usa este
  2. OpenVINO CPU:  15ms
  3. Ollama CPU:    20ms
```

## Instalação Autônoma (Non-interactive)

Para scripts/CI/CD:

```bash
# Instala o melhor provider para o hardware atual
openforge provider install --auto

# Específico
openforge provider install ollama --non-interactive
openforge provider install openvino --non-interactive --model-path ./models

# Init completo (auto-detect + config + modelo recomendado)
openforge init --auto
```

## Comandos `openforge init`

```bash
openforge init                  # Interativo: pergunta preferências
openforge init --auto           # Tudo automático: detecta + configura + instala
openforge init --provider ollama # Init forçando Ollama
```

O init:
1. Detecta hardware (CPU, GPU, NPU)
2. Detecta runtimes instalados
3. Instala runtime recomendado (se `--auto`)
4. Gera `config.yaml` otimizado
5. Baixa modelo recomendado (se solicitado)
6. Testa inferência (health check)

## Estrutura de Diretórios

```
internal/pm/
├── pm.go              # ProviderManager (orquestrador)
├── types.go           # Tipos: ProviderType, ProviderStatus, ProviderInfo
├── discovery.go       # Auto-detecção de providers instalados
├── install.go         # Instalação guiada e automática
├── config.go          # Geração de configuração
├── chain.go           # Lógica de chain/resolution
├── health.go          # Health check de providers
├── providers/         # Implementações
│   ├── openvino.go
│   ├── ollama.go
│   ├── llamacpp.go
│   ├── vllm.go
│   └── lmstudio.go
└── installers/        # Scripts de instalação por SO
    ├── ollama.go
    ├── openvino.go
    ├── llamacpp.go
    └── vllm.go
```
