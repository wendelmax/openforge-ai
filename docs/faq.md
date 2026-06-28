# FAQ

## General

### Does this work without internet?

**Yes.** OpenForge is designed for 100% offline operation. After installation and model download, no internet connection is required.

### What hardware do I need?

Any Intel CPU (Core i 6th gen+ or Xeon). GPU (Iris Xe, Arc) and NPU (AI Boost) are optional but recommended for better performance.

### Does it work on AMD?

CPU-only. OpenVINO runs on AMD CPUs via the CPU plugin. GPU and NPU are Intel-specific.

### Does it work on Apple Silicon?

CPU-only via OpenVINO's ARM plugin. GPU and NPU are not supported on Apple hardware.

### What is OpenVINO?

OpenVINO (Open Visual Inference & Neural Network Optimization) is Intel's open-source toolkit for optimizing and deploying AI models on Intel hardware.

## Installation

### How do I install OpenVINO?

See [Installation](getting-started/installation.md) for platform-specific instructions.

### Can I use Docker?

Yes. Official images are available at `docker pull openforge/openforge`.

### Do I need a GPU?

No. Everything works on CPU. GPU and NPU provide faster inference.

## Models

### Where do I get models?

Use `openforge model pull <name>` to download from the official model zoo, or convert your own using Optimum Intel.

### What format are the models?

OpenVINO IR format: `.xml` (graph) + `.bin` (weights).

### Can I use GGUF models?

Not directly. Convert to OpenVINO IR first using the `openforge model convert` skill.

### How much disk space do models need?

| Model | Disk Space |
|-------|:----------:|
| Phi-3 Mini (INT4) | ~1.5 GB |
| Llama 3.2 3B (INT4) | ~1.8 GB |
| BGE Small | ~0.1 GB |
| Mistral 7B (INT4) | ~4.0 GB |

## Performance

### How fast is it compared to Ollama?

On Intel hardware with GPU/NPU, OpenForge is typically **20-40% faster** due to direct OpenVINO integration without the Ollama server overhead.

### What is the NPU advantage?

NPU excels at:
- **Embedding**: 3x faster than CPU, 1.5x faster than GPU
- **Small LLMs** (< 3B): 2x faster than CPU
- **Power efficiency**: 5x better perf/watt than GPU

### Can I run multiple models simultaneously?

Not yet. Multi-model serving is planned for v1.0.

## Usage

### Does it work with OpenCode?

**Yes.** OpenForge is designed as a native OpenCode provider. Configure it as:
```json
{"provider": "openforge", "endpoint": "http://localhost:9090/v1"}
```

### Does it work with the OpenAI SDK?

**Yes.** OpenForge API is OpenAI-compatible. Any OpenAI client can point to `http://localhost:9090/v1`.

### What is a Skill?

A skill is a reusable YAML pipeline that combines multiple AI steps (prompts, embeddings, reranking) into a single command. See [Creating Skills](skills/creating-skills.md).

## Troubleshooting

### OpenForge starts but models won't load

1. Check model path: `ls ./models/`
2. Verify OpenVINO installation: `ovc --version`
3. Check available devices: `openforge devices`
4. Check logs: `openforge serve --verbose`

### GPU/NPU not detected

1. Install GPU/NPU drivers from Intel
2. Verify with OpenVINO: `benchmark_app -d GPU`
3. Restart OpenForge

### Out of memory errors

- Use INT4 quantized models instead of FP16
- Close other GPU/NPU applications
- Reduce `max_tokens` in your requests

### Performance is slow

- Check device: `openforge devices` — are you on CPU?
- Check model quantization: prefer INT4
- Run a benchmark: `openforge benchmark --model <name>`

## Contributing

### How can I contribute?

See [Contributing](contributing.md). We welcome code, docs, tests, and skills.

### Can I create my own Skill?

Yes. Skills are YAML files. See [Creating Skills](skills/creating-skills.md).

### Can I create my own Provider?

Yes. Implement the Provider interface. See [Providers Guide](providers-guide.md).
