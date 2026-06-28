# Model Zoo

All models listed here are tested and supported. Models are in OpenVINO IR format (`.xml` + `.bin`).

## LLM (Large Language Models)

| Model | Size | Precision | RAM | NPU | GPU | Context | License |
|-------|:----:|:---------:|:---:|:---:|:---:|:-------:|:--------|
| Phi-3 Mini | 3.8B | INT4 | 2.1 GB | ✅ | ✅ | 4K | MIT |
| Phi-3 Mini | 3.8B | FP16 | 7.2 GB | — | ✅ | 4K | MIT |
| Llama 3.2 3B | 3B | INT4 | 1.8 GB | ✅ | ✅ | 8K | Llama 3 |
| Llama 3.2 3B | 3B | FP16 | 6.1 GB | — | ✅ | 8K | Llama 3 |
| Llama 3.1 8B | 8B | INT4 | 4.5 GB | — | ✅ | 128K | Llama 3 |
| Mistral 7B | 7B | INT4 | 4.2 GB | — | ✅ | 32K | Apache 2.0 |
| Qwen2 0.5B | 0.5B | INT4 | 0.4 GB | ✅ | ✅ | 32K | Apache 2.0 |
| Qwen2 0.5B | 0.5B | FP16 | 1.0 GB | — | ✅ | 32K | Apache 2.0 |
| Qwen2 1.5B | 1.5B | INT4 | 1.0 GB | ✅ | ✅ | 32K | Apache 2.0 |
| CodeGemma 2B | 2B | INT4 | 1.2 GB | ✅ | ✅ | 8K | Gemma |

## Embedding

| Model | Dims | RAM | NPU | GPU | License |
|-------|:----:|:---:|:---:|:---:|:--------|
| BGE Small EN v1.5 | 384 | 0.1 GB | ✅ | ✅ | MIT |
| BGE Base EN v1.5 | 768 | 0.3 GB | ✅ | ✅ | MIT |
| BGE Large EN v1.5 | 1024 | 1.0 GB | — | ✅ | MIT |
| BGE M3 | 1024 | 1.2 GB | — | ✅ | MIT |
| BGE Reranker v2 | 1 | 0.5 GB | ✅ | ✅ | MIT |
| BGE Reranker v2 M3 | 1 | 1.1 GB | — | ✅ | MIT |

## Vision (Future)

| Model | Size | RAM | NPU | GPU |
|-------|:----:|:---:|:---:|:---:|
| YOLOv8 (planned) | — | — | — | — |
| ResNet-50 (planned) | — | — | — | — |

## Audio (Future)

| Model | Size | RAM | NPU | GPU |
|-------|:----:|:---:|:---:|:---:|
| Whisper Base (planned) | — | — | — | — |
| Whisper Small (planned) | — | — | — | — |

## Download a Model

```bash
# Official model zoo
openforge model pull phi-3-mini-int4
openforge model pull bge-small-en-v1.5

# List available models
openforge model list-remote

# List local models
openforge model list-local

# Custom model (add to models/ directory)
# The model directory should contain:
#   model.xml       (OpenVINO IR graph)
#   model.bin       (OpenVINO IR weights)
#   tokenizer.json  (HuggingFace tokenizer)
```

## Convert Your Own

Use [Optimum Intel](https://github.com/huggingface/optimum-intel) to convert HuggingFace models:

```bash
pip install optimum[openvino]

optimum-cli export openvino \
  --model microsoft/phi-3-mini-4k-instruct \
  --weight-format int4 \
  phi-3-mini-int4/
```
