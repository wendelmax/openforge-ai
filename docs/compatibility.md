# Compatibility

## Operating Systems

| OS | CPU | GPU | NPU | Status |
|:---|:---:|:---:|:---:|:------|
| Ubuntu 22.04+ | ✅ Full | ✅ Full | ✅ Full | Tier 1 |
| Ubuntu 24.04+ | ✅ Full | ✅ Full | ✅ Full | Tier 1 |
| Fedora 40+ | ✅ Full | ✅ Full | ✅ Full | Tier 1 |
| Windows 10/11 | ✅ Full | ✅ Full | ✅ Full | Tier 2 |
| macOS 14+ (Apple Silicon) | ✅ CPU only | — | — | Tier 3 |
| macOS 14+ (Intel) | ✅ Full | ✅ Full | — | Tier 3 |
| Debian 12+ | ✅ Full | ✅ Full | ✅ Partial | Tier 2 |
| Arch Linux | ✅ Full | ✅ Full | ✅ Partial | Tier 2 |
| RHEL 9+ | ✅ Full | ✅ Full | — | Tier 3 |

**Tier 1**: Officially tested in CI, full support  
**Tier 2**: Community-tested, best-effort support  
**Tier 3**: Experimental, may have limitations

## Hardware

| Vendor | CPU | GPU | NPU |
|:-------|:---:|:---:|:---:|
| Intel Core (12th gen+) | ✅ | ✅ | ✅ |
| Intel Core Ultra (Meteor Lake+) | ✅ | ✅ | ✅ |
| Intel Xeon (4th gen+) | ✅ | — | — |
| Intel Iris Xe | — | ✅ | — |
| Intel Arc A-Series | — | ✅ | — |
| Intel AI Boost (NPU) | — | — | ✅ |
| AMD Ryzen | ✅ | — | — |
| AMD Threadripper | ✅ | — | — |
| Apple M1/M2/M3/M4 | ✅ | — | — |

## OpenVINO Versions

| OpenVINO | OpenForge | Status |
|:---------|:---------:|:------:|
| 2025.0 | ≥ 0.1.0 | ✅ Supported |
| 2024.6 | ≥ 0.1.0 | ✅ Supported |
| 2024.5 | ≥ 0.1.0 | ✅ Supported |
| 2024.4 | — | ❌ End of life |
| 2024.3 | — | ❌ End of life |
| 2023.x | — | ❌ Not supported |

## Provider Compatibility

| Provider | Status | Notes |
|:---------|:------:|:------|
| OpenVINO | ✅ Native | Default and only provider in v0.x |
| Plugin API | 🔧 Planned | Custom providers via plugin system |

## Client Compatibility

| Client | Chat | Completion | Embed | Rerank | Stream |
|:-------|:----:|:----------:|:-----:|:------:|:------:|
| OpenCode | ✅ | ✅ | ✅ | — | ✅ |
| Superpowers | ✅ | ✅ | ✅ | — | ✅ |
| OpenAI Python SDK | ✅ | ✅ | ✅ | — | ✅ |
| curl | ✅ | ✅ | ✅ | ✅ | ✅ |
| VS Code Extension | 🔧 | 🔧 | — | — | 🔧 |
| IntelliJ Plugin | 🔧 | 🔧 | — | — | 🔧 |

## Framework Support (for Skills)

| Framework | Status |
|:----------|:------:|
| Go | ✅ |
| Java / Spring | ✅ |
| Node.js / NestJS | ✅ |
| Python / Django | ✅ |
| Rust | 🔧 |
| .NET | 🔧 |
| Kubernetes | 🔧 |
