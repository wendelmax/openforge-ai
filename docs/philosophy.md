# OpenForge Manifesto

A filosofia que guia cada decisão do projeto.

## Local First

AI assistants should run on your machine, not in someone else's cloud. Local execution means:
- **Privacy**: your code and prompts never leave your hardware
- **Latency**: no network round-trips, no queueing
- **Reliability**: works in a datacenter, on a plane, or in a bunker
- **Cost**: no API fees, no per-token pricing

## OpenVINO First

We bet on OpenVINO because it's the best runtime for Intel hardware. This means:
- One runtime, one optimization path, one set of guarantees
- Native support for CPU, GPU, and NPU
- No abstraction layers between the model and the silicon
- Continuous optimization from Intel's compiler engineers

## Offline by Default

No telemetry. No analytics. No "phone home." If a feature requires internet, it's opt-in and clearly documented.

## Community Driven

- Spec-driven development with open ADRs
- Transparent roadmap and decision-making
- Skills shared as files, not as a marketplace
- Plugin API for community extensions
- License: Apache 2.0 — free forever

## Modular

- Providers are swappable (one today, many tomorrow)
- Skills are composable YAML pipelines
- Plugins can extend any layer
- Cache is configurable (memory, SQLite, custom)

## Extensible

The plugin system means the community can extend OpenForge without waiting for us:
- New providers (ONNX, llama.cpp, custom hardware)
- New skills (any language, any framework)
- New integrations (IDEs, CI/CD, editors)

## Reproducible

- Same model, same prompt, same device → same output
- Benchmarks are versioned and comparable
- Config as code (YAML + env vars)
- Docker images for consistent deployment

## Transparent

- Every architectural decision documented as an ADR
- Open metrics and benchmarks
- Public roadmap with clear priorities
- No dark patterns, no vendor lock-in

---

## What This Means for You

| Concern | Our Answer |
|---------|------------|
| "Can I trust this with my code?" | Yes — it never leaves your machine |
| "Will I be locked in?" | No — open source, open formats, open API |
| "Can I extend it?" | Yes — Skills and Plugins are designed for extension |
| "Will it work on my hardware?" | Yes — CPU always works, GPU/NPU if available |
| "Is it production ready?" | We're working toward it — see the Roadmap |
