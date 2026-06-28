# Installation

## Prerequisites

- **CPU**: Intel Core (6th gen+) or compatible x86_64
- **GPU** (optional): Intel Iris Xe, Intel Arc
- **NPU** (optional): Intel AI Boost (Meteor Lake+)
- **OS**: Linux (Tier 1), Windows (Tier 2), macOS (Tier 3)
- **OpenVINO**: 2025.x runtime

## Quick Install (Linux / macOS)

```bash
curl -fsSL https://openforge.ai/install.sh | bash
```

This installs the latest binary for your architecture.

## Manual Install

### 1. Download

```bash
# Linux amd64
wget https://github.com/openforge-ai/openforge/releases/latest/download/openforge-linux-amd64.tar.gz
tar -xzf openforge-linux-amd64.tar.gz
sudo mv openforge /usr/local/bin/
```

### 2. Install OpenVINO Runtime

```bash
# Linux (APT)
wget https://apt.repos.intel.com/intel-gpg-keys/GPG-PUB-KEY-INTEL-SW-PRODUCTS.PUB
sudo apt-key add GPG-PUB-KEY-INTEL-SW-PRODUCTS.PUB
sudo add-apt-repository "deb https://apt.repos.intel.com/openvino/2025 ubuntu22 main"
sudo apt update
sudo apt install openvino-2025

# Verify
ovc --version
```

### 3. Verify

```bash
openforge version
openforge devices
```

## Docker

```bash
docker pull openforge/openforge:latest

docker run -d \
  -p 9090:9090 \
  -v /path/to/models:/models \
  openforge/openforge:latest
```

## Build from Source

```bash
git clone https://github.com/openforge-ai/openforge.git
cd openforge
go build -o build/openforge ./cmd/server
```

## Post-Install

Download your first model:

```bash
openforge model pull llama-3.2-3b
```

Start the server:

```bash
openforge serve
```
