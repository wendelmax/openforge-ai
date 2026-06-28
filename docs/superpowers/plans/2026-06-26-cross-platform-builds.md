# Cross-Platform Distribution Builds Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generate installable distribution archives for Linux (amd64 + arm64), Windows (amd64 + arm64), and Darwin (arm64) with OpenVINO runtime bundled for CGO-capable platforms.

**Architecture:** Use goreleaser with Docker builder images for CGO-enabled builds. Linux amd64 CGO builds inside a Docker container with OpenVINO 2026.2.1 SDK. Windows amd64 CGO builds via goreleaser's Docker builder with MinGW cross-compiler + Windows OpenVINO headers/libs. All other targets (arm64, darwin) use CGO_ENABLED=0 stub builds. Archives bundle OpenVINO .so/.dll runtime files alongside the binary.

**Tech Stack:** Go 1.26, Goreleaser v2, Docker, OpenVINO 2026.2.1, MinGW-w64 (for Windows cross-build)

## Global Constraints

- All existing tests must pass
- CGO_ENABLED=0 builds must work on all platforms for stub mode
- CGO_ENABLED=1 builds must work on linux/amd64 and windows/amd64
- OpenVINO runtime libs must be bundled in the archive for CGO targets
- Non-CGO targets (arm64, darwin) must produce a working stub binary
- Docker builder image must be reproducible and version-pinned
- CI must verify all build targets on every push
- Archive naming: `openforge-{os}-{arch}.tar.gz` (unix) / `.zip` (windows)

---

### Task 1: Create Docker builder image for CGO builds

**Files:**
- Create: `build/Dockerfile.builder` (multi-stage builder with OpenVINO SDK + MinGW)
- Create: `build/builder-env.sh` (env var setup for the builder)
- Review: `internal/provider/openvino/binding.go` (confirm CGO flags)

**Interfaces:**
- Consumes: OpenVINO 2026.2.1 archive URL, MinGW packages
- Produces: Docker image `openforge/builder:go1.26-ov2026.2.1` with all build dependencies

**Why a Docker builder:** Goreleaser supports `builds[].builder: docker` which runs the build inside a container and extracts the binary. We create a single image that has:
- Go 1.26.0 (official binary install)
- OpenVINO 2026.2.1 C SDK (headers + .so + .a for Linux, headers + .dll.a + .h for Windows cross)
- MinGW-w64 for Windows cross-compilation
- pkg-config

- [ ] **Step 1: Determine OpenVINO archive URLs**

Linux x86_64 archive URL (from WSL setup):
```
https://storage.openvinotoolkit.org/repositories/openvino/packages/2026.2.1/linux/openvino_toolkit_ubuntu24_2026.2.1.21919.ede283a88e3_x86_64.tar.gz
```

Windows x86_64 archive URL (from NuGet or official release):
```
https://storage.openvinotoolkit.org/repositories/openvino/packages/2026.2.1/windows/openvino_toolkit_windows_2026.2.1.21919.ede283a88e3_x86_64.zip
```

Set these as build args so they can be updated when OpenVINO releases new versions.

- [ ] **Step 2: Write build/Dockerfile.builder**

```dockerfile
# Stage 1: Base with Go + build tools
FROM ubuntu:24.04 AS base

ARG OV_VERSION=2026.2.1.21919.ede283a88e3
ARG OV_LINUX_URL=https://storage.openvinotoolkit.org/repositories/openvino/packages/2026.2.1/linux/openvino_toolkit_ubuntu24_${OV_VERSION}_x86_64.tar.gz
ARG OV_WIN_URL=https://storage.openvinotoolkit.org/repositories/openvino/packages/2026.2.1/windows/openvino_toolkit_windows_2026.2.1.21919.ede283a88e3_x86_64.zip

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    xz-utils \
    zip \
    unzip \
    pkg-config \
    mingw-w64 \
    && rm -rf /var/lib/apt/lists/*

# Go 1.26
RUN curl -fsSL https://go.dev/dl/go1.26.0.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH=/usr/local/go/bin:$PATH

# Linux OpenVINO SDK
RUN mkdir -p /opt/ov-linux && \
    curl -fsSL "$OV_LINUX_URL" | tar -C /opt/ov-linux -xz --strip-components=1

# Windows OpenVINO SDK (for cross-compilation)
RUN mkdir -p /opt/ov-win && \
    curl -fsSL -o /tmp/ov-win.zip "$OV_WIN_URL" && \
    unzip -q /tmp/ov-win.zip -d /opt/ov-win && \
    mv /opt/ov-win/*/* /opt/ov-win/ && \
    rm /tmp/ov-win.zip

# pkg-config files for Linux
RUN mkdir -p /opt/pkgconfig && \
    cat > /opt/pkgconfig/openvino.pc << 'EOF'
prefix=/opt/ov-linux
exec_prefix=${prefix}
libdir=${exec_prefix}/runtime/lib/intel64
includedir=${prefix}/runtime/include

Name: OpenVINO
Description: Intel OpenVINO Toolkit
Version: 2026.2.1
Libs: -L${libdir} -lopenvino_c -lopenvino
Cflags: -I${includedir}
EOF

# pkg-config files for Windows (cross)
RUN mkdir -p /opt/pkgconfig-win && \
    cat > /opt/pkgconfig-win/openvino.pc << 'EOF'
prefix=/opt/ov-win
exec_prefix=${prefix}
libdir=${exec_prefix}/runtime/lib/intel64
includedir=${prefix}/runtime/include

Name: OpenVINO
Description: Intel OpenVINO Toolkit
Version: 2026.2.1
Libs: -L${libdir} -lopenvino_c -lopenvino
Cflags: -I${includedir}
EOF

ENV PKG_CONFIG_PATH=/opt/pkgconfig
ENV OV_LINUX_DIR=/opt/ov-linux
ENV OV_WIN_DIR=/opt/ov-win
WORKDIR /build
```

This image is used as the builder and referenced in the goreleaser Docker builder config.

- [ ] **Step 3: Write build/builder-env.sh**

Helper script for local builds (non-goreleaser):

```bash
#!/bin/bash
# Source this file to set up build environment for CGO builds
# Usage: source build/builder-env.sh [linux|windows]

export PKG_CONFIG_PATH=/opt/pkgconfig
export CGO_ENABLED=1

if [ "$1" = "windows" ]; then
    export GOOS=windows
    export GOARCH=amd64
    export CC=x86_64-w64-mingw32-gcc
    export PKG_CONFIG_PATH=/opt/pkgconfig-win
    export CGO_CFLAGS="-I/opt/ov-win/runtime/include"
    export CGO_LDFLAGS="-L/opt/ov-win/runtime/lib/intel64 -lopenvino_c -lopenvino"
else
    export GOOS=linux
    export GOARCH=amd64
    export CC=gcc
    export CGO_CFLAGS="-I/opt/ov-linux/runtime/include"
    export CGO_LDFLAGS="-L/opt/ov-linux/runtime/lib/intel64 -lopenvino_c -lopenvino"
fi
```

- [ ] **Step 4: Build and test the builder image**

```bash
docker build -t openforge/builder:go1.26-ov2026.2.1 -f build/Dockerfile.builder .
```

Verify the image works:
```bash
docker run --rm -v $(pwd):/build openforge/builder:go1.26-ov2026.2.1 \
  bash -c "source build/builder-env.sh linux && go build ./internal/provider/openvino/..."
```

---

### Task 2: Update goreleaser config for CGO builds

**Files:**
- Modify: `.goreleaser.yaml`
- Modify: `internal/provider/openvino/binding.go` (add pkg-config fallback or env var fallback)

**Interfaces:**
- Consumes: Task 1 builder image
- Produces: Goreleaser config with CGO + stub builds, archive definitions

- [ ] **Step 1: Split builds into CGO and non-CGO targets**

```yaml
version: 2

project_name: openforge

before:
  hooks:
    - go mod tidy

builds:
  # --- CGO builds (require OpenVINO SDK) ---

  # Linux amd64 with OpenVINO
  - id: openforge-linux-amd64-cgo
    main: ./cmd/openforge
    binary: openforge-linux-amd64
    builder: docker
    docker:
      image: openforge/builder:go1.26-ov2026.2.1
      hide_pre: true
    env:
      - CGO_ENABLED=1
      - GOOS=linux
      - GOARCH=amd64
      - PKG_CONFIG_PATH=/opt/pkgconfig
    ldflags:
      - -s -w
      - -X github.com/openforge-ai/openforge/cmd/openforge/cmd.Version={{ .Version }}
      - -X github.com/openforge-ai/openforge/cmd/openforge/cmd.Commit={{ .Commit }}
      - -X github.com/openforge-ai/openforge/cmd/openforge/cmd.Date={{ .Date }}
    goos:
      - linux
    goarch:
      - amd64

  # Windows amd64 with OpenVINO (cross-compiled via MinGW)
  - id: openforge-windows-amd64-cgo
    main: ./cmd/openforge
    binary: openforge-windows-amd64
    builder: docker
    docker:
      image: openforge/builder:go1.26-ov2026.2.1
      hide_pre: true
    env:
      - CGO_ENABLED=1
      - GOOS=windows
      - GOARCH=amd64
      - CC=x86_64-w64-mingw32-gcc
      - PKG_CONFIG_PATH=/opt/pkgconfig-win
    ldflags:
      - -s -w
      - -X github.com/openforge-ai/openforge/cmd/openforge/cmd.Version={{ .Version }}
      - -X github.com/openforge-ai/openforge/cmd/openforge/cmd.Commit={{ .Commit }}
      - -X github.com/openforge-ai/openforge/cmd/openforge/cmd.Date={{ .Date }}
    goos:
      - windows
    goarch:
      - amd64

  # --- CGO_ENABLED=0 stub builds (all platforms) ---

  - id: openforge-stub
    main: ./cmd/openforge
    binary: openforge
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w
      - -X github.com/openforge-ai/openforge/cmd/openforge/cmd.Version={{ .Version }}
      - -X github.com/openforge-ai/openforge/cmd/openforge/cmd.Commit={{ .Commit }}
      - -X github.com/openforge-ai/openforge/cmd/openforge/cmd.Date={{ .Date }}
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: linux
        goarch: amd64    # handled by CGO build
      - goos: windows
        goarch: amd64    # handled by CGO build
      - goos: windows
        goarch: arm64    # not supported
      - goos: darwin
        goarch: amd64    # not supported
```

Note: The `ignore` excludes linux/amd64 and windows/amd64 from stub builds since they have dedicated CGO builds. But for users who want stub-only builds, we should keep those. Actually, we should NOT exclude them — we want both CGO and stub archives available. But the binary names would conflict. Better approach: name the CGO binaries differently, or keep them as alternatives.

Actually, a cleaner approach is:
- CGO builds produce `openforge-cgo-linux-amd64` and `openforge-cgo-windows-amd64`
- Stub builds produce `openforge-linux-amd64`, `openforge-windows-amd64`, etc.
- The "latest" / default is the CGO version for platforms that support it

But goreleaser builds one binary per `goos`/`goarch` combination. If we have two builds producing the same os/arch, they'd need different binary names (which they do, via `id`).

Let me use a simpler approach: the CGO build ID is `openforge-cgo` and produces a binary named `openforge` (same as stub), but for different os/arch combos. When both CGO and stub exist for the same os/arch, the later one wins. So let me make CGO builds explicitly named.

Actually, the simplest pattern: use separate archive groups and name templates that distinguish CGO from stub.

- [ ] **Step 2: Configure archives with OpenVINO runtime bundles**

```yaml
archives:
  # Archives for CGO Linux builds — include OpenVINO .so files
  - id: cgo-linux
    builds:
      - openforge-linux-amd64-cgo
    formats: tar.gz
    name_template: "{{ .ProjectName }}-{{ .Os }}-{{ .Arch }}-with-openvino"
    files:
      - README.md
      - LICENSE
      - src: "/opt/ov-linux/runtime/lib/intel64/libopenvino_c.so*"
        dst: "lib/"
        strip_parent: true
      - src: "/opt/ov-linux/runtime/lib/intel64/libopenvino.so*"
        dst: "lib/"
        strip_parent: true
      - src: "/opt/ov-linux/runtime/lib/intel64/libopenvino_intel_cpu_plugin.so"
        dst: "lib/"
        strip_parent: true

  # Archives for CGO Windows builds — include OpenVINO .dll files
  - id: cgo-windows
    builds:
      - openforge-windows-amd64-cgo
    formats: zip
    name_template: "{{ .ProjectName }}-{{ .Os }}-{{ .Arch }}-with-openvino"
    files:
      - README.md
      - LICENSE
      - src: "/opt/ov-win/runtime/lib/intel64/openvino_c.dll"
        dst: ""
      - src: "/opt/ov-win/runtime/lib/intel64/openvino.dll"
        dst: ""
      - src: "/opt/ov-win/runtime/lib/intel64/openvino_intel_cpu_plugin.dll"
        dst: ""

  # Archives for stub builds — no OpenVINO runtime needed
  - id: stub
    builds:
      - openforge-stub
    formats: tar.gz
    name_template: "{{ .ProjectName }}-{{ .Os }}-{{ .Arch }}"
    files:
      - README.md
      - LICENSE
```

- [ ] **Step 3: Test the goreleaser config**

```bash
# Snapshot build to test
goreleaser build --snapshot --clean
```

Verify:
- CGO builds produce binaries linked against OpenVINO
- Stub builds produce standalone binaries
- Archives contain expected files

- [ ] **Step 4: Update binding.go for fallback CGO flags**

The current `binding.go` uses `#cgo LDFLAGS: -lopenvino_c -lopenvino`. For CGO builds inside Docker, pkg-config is already set up. But the Docker builder also provides CGO_CFLAGS/CGO_LDFLAGS via env vars. The `#cgo LDFLAGS` in binding.go should remain as just `-lopenvino_c -lopenvino` (no `pkg-config` directive) since we supply include/lib paths via env vars.

Verify `internal/provider/openvino/binding.go` — it should NOT have `#cgo pkg-config: openvino`. If it does, remove it. The linker flags `-lopenvino_c -lopenvino` are fine because the Docker build env provides the lib search path.

---

### Task 3: Update production Dockerfile

**Files:**
- Modify: `Dockerfile`

**Interfaces:**
- Consumes: CGO binary from builder image
- Produces: Production Docker image with OpenVINO runtime

- [ ] **Step 1: Restructure Dockerfile**

```dockerfile
# Stage 1: Build with OpenVINO SDK
FROM openforge/builder:go1.26-ov2026.2.1 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN source build/builder-env.sh linux && \
    go build -ldflags="-s -w -X github.com/openforge-ai/openforge/cmd/openforge/cmd.Version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -o openforge ./cmd/openforge

# Stage 2: Runtime with OpenVINO shared libs
FROM ubuntu:24.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

RUN adduser -D -h /home/openforge openforge 2>/dev/null || \
    useradd -d /home/openforge -m openforge

# Copy OpenVINO runtime libraries
COPY --from=builder /opt/ov-linux/runtime/lib/intel64/libopenvino_c.so* /usr/local/lib/
COPY --from=builder /opt/ov-linux/runtime/lib/intel64/libopenvino.so* /usr/local/lib/
COPY --from=builder /opt/ov-linux/runtime/lib/intel64/libopenvino_intel_cpu_plugin.so /usr/local/lib/
RUN ldconfig

COPY --from=builder /build/openforge /usr/local/bin/openforge

USER openforge
WORKDIR /home/openforge

EXPOSE 9090
ENTRYPOINT ["openforge"]
CMD ["serve"]
```

- [ ] **Step 2: Build and test the Docker image**

```bash
docker build -t openforge:test .
docker run --rm openforge:test --help
```

Expected: OpenVINO runtime loads, device detection works inside container.

---

### Task 4: Update GitHub Actions CI

**Files:**
- Modify: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: Task 1 builder image, Task 2 goreleaser config
- Produces: CI pipeline that validates all builds

- [ ] **Step 1: Update Go version to 1.26**

Change `go-version: "1.23"` to `go-version: "1.26"` in all CI job steps.

- [ ] **Step 2: Update CGO test job for OpenVINO 2026.2.1**

```yaml
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        cgo: ["0", "1"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - name: Install OpenVINO (CGO=1)
        if: matrix.cgo == '1'
        run: |
          wget -qO- https://storage.openvinotoolkit.org/repositories/openvino/packages/2026.2.1/linux/openvino_toolkit_ubuntu24_2026.2.1.21919.ede283a88e3_x86_64.tar.gz | tar -C /opt -xz --strip-components=1
          echo "PKG_CONFIG_PATH=/opt/runtime/lib/pkgconfig" >> $GITHUB_ENV
          echo "LD_LIBRARY_PATH=/opt/runtime/lib/intel64" >> $GITHUB_ENV
      - name: Test
        run: CGO_ENABLED=${{ matrix.cgo }} go test ./... -count=1 -coverprofile=coverage.out
```

Note: The `PKG_CONFIG_PATH` method relies on OpenVINO's own pkgconfig files. Our binding.go might need a `#cgo pkg-config: openvino` line for this to work. Check if the OpenVINO archive provides `.pc` files. If not, use env vars instead.

Actually, looking at our current WSL setup, we created a custom pkgconfig. The official OpenVINO archive at `/tmp/ov/openvino_toolkit_ubuntu24_2026.2.1.21919.ede283a88e3_x86_64/runtime/lib/pkgconfig` should have .pc files. Let me verify in the plan that they exist or we provide fallback.

- [ ] **Step 3: Add CGO build validation job**

Add a job that builds the Docker builder image and validates CGO builds:

```yaml
  build-cgo:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build builder image
        run: docker build -t openforge/builder -f build/Dockerfile.builder .
      - name: Build CGO binary
        run: |
          docker run --rm -v $(pwd):/build openforge/builder \
            bash -c "source build/builder-env.sh linux && go build -o /build/openforge-cgo-linux-amd64 ./cmd/openforge"
      - name: Verify binary
        run: file openforge-cgo-linux-amd64
      - name: Upload CGO binary
        uses: actions/upload-artifact@v4
        with:
          name: openforge-cgo-linux-amd64
          path: openforge-cgo-linux-amd64
```

- [ ] **Step 4: Add release workflow**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - name: Log in to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      - name: Build builder image
        run: docker build -t openforge/builder -f build/Dockerfile.builder .
      - name: Run goreleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

### Task 5: Add build documentation and local build targets

**Files:**
- Modify: `Taskfile.yml` (add dist targets)
- Modify: `docs/build.md` (document build process)

**Interfaces:**
- Consumes: All prior tasks
- Produces: Developer-friendly build workflow

- [ ] **Step 1: Add dist targets to Taskfile.yml**

```yaml
  dist:docker-builder:
    desc: Build the Docker builder image for CGO builds
    cmds:
      - docker build -t openforge/builder:go1.26-ov2026.2.1 -f build/Dockerfile.builder .

  dist:linux-cgo:
    desc: Build CGO binary for Linux amd64 inside Docker
    cmds:
      - docker run --rm -v $(pwd):/build openforge/builder:go1.26-ov2026.2.1 \
          bash -c "source build/builder-env.sh linux && go build -ldflags='-s -w' -o /build/build/openforge-linux-amd64-cgo ./cmd/openforge"
    deps: [dist:docker-builder]

  dist:windows-cgo:
    desc: Build CGO binary for Windows amd64 (cross-compile inside Docker)
    cmds:
      - docker run --rm -v $(pwd):/build openforge/builder:go1.26-ov2026.2.1 \
          bash -c "source build/builder-env.sh windows && go build -ldflags='-s -w' -o /build/build/openforge-windows-amd64-cgo.exe ./cmd/openforge"
    deps: [dist:docker-builder]

  dist:stub:
    desc: Build stub binaries for all platforms
    cmds:
      - task: build:linux
      - task: build:windows
      - task: build:darwin

  dist:
    desc: Build all distribution binaries (CGO + stub)
    deps: [dist:stub, dist:linux-cgo, dist:windows-cgo]

  release:
    desc: Create a full release with goreleaser (requires git tag)
    cmds:
      - goreleaser release --clean
```

- [ ] **Step 2: Write build documentation**

In `docs/build.md`:

```markdown
# Building OpenForge

## Prerequisites
- Go 1.26+
- Docker (for CGO builds)
- Goreleaser (for releases)

## Quick Build (Stub Mode)
```bash
go build -o openforge ./cmd/openforge
```

## CGO Builds (with OpenVINO)

### Using Docker (recommended)
```bash
# Build the builder image
docker build -t openforge/builder -f build/Dockerfile.builder .

# Build Linux CGO binary
docker run --rm -v $(pwd):/build openforge/builder \
  bash -c "source build/builder-env.sh linux && go build -o /build/openforge ./cmd/openforge"

# Build Windows CGO binary (cross-compile)
docker run --rm -v $(pwd):/build openforge/builder \
  bash -c "source build/builder-env.sh windows && go build -o /build/openforge.exe ./cmd/openforge"
```

### Using Taskfile
```bash
task dist:linux-cgo    # Linux amd64 with OpenVINO
task dist:windows-cgo  # Windows amd64 with OpenVINO
task dist              # All platforms
```

## Release
```bash
git tag v0.1.0
goreleaser release --clean
```

## Supported Platforms
| Platform | CGO (OpenVINO) | Stub |
|----------|----------------|------|
| Linux amd64 | ✅ Full support | ✅ |
| Linux arm64 | ❌ | ✅ |
| Windows amd64 | ✅ Full support | ✅ |
| Windows arm64 | ❌ | ✅ |
| Darwin arm64 | ❌ | ✅ |
```

- [ ] **Step 3: Remove cross-compile CGO from .goreleaser.yaml and just do Linux CGO**

Actually, reconsider: Windows CGO cross-compilation with MinGW + OpenVINO is complex and may not work reliably. Let me adjust the goreleaser config to only do Linux CGO, and document Windows CGO as a manual/native build step.

Actually, I should test whether the MinGW cross-compilation works at all. Let me plan to test it during implementation. The task should include verification.

---

### Task 6: Verify and test the full pipeline

**Files:**
- None (verification only)

**Interfaces:**
- Consumes: All prior tasks

- [ ] **Step 1: Verify Linux CGO binary works**

```bash
# Build the binary
docker run --rm -v $(pwd):/build openforge/builder \
  bash -c "source build/builder-env.sh linux && go build -o /build/build/openforge ./cmd/openforge"

# Run it and check device detection
file build/openforge
ldd build/openforge | grep openvino
```

- [ ] **Step 2: Verify Windows CGO binary works (cross-compile)**

```bash
docker run --rm -v $(pwd):/build openforge/builder \
  bash -c "source build/builder-env.sh windows && go build -o /build/build/openforge.exe ./cmd/openforge"

file build/openforge.exe
# Check it's a PE32+ executable
```

Note: The Windows binary can't be run on Linux, but we verify it links correctly.

- [ ] **Step 3: Verify stub binaries for all platforms**

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o build/openforge-linux-arm64 ./cmd/openforge
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -o build/openforge-windows-arm64.exe ./cmd/openforge
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o build/openforge-darwin-arm64 ./cmd/openforge
```

All should produce working binaries with stub OpenVINO.

- [ ] **Step 4: Run all existing tests**

```bash
CGO_ENABLED=0 go test ./... -count=1
CGO_ENABLED=1 go test ./internal/provider/openvino/... -count=1   # or skip if no OpenVINO
```

- [ ] **Step 5: Verify goreleaser snapshot**

```bash
goreleaser build --snapshot --clean
```

Check that:
- Archives are created for each target
- CGO archives include OpenVINO runtime libs
- Stub archives are standalone

---

### Self-Review

1. **Spec coverage:** Plan covers Docker builder image, goreleaser CGO/stub split, archive bundling, CI pipeline, Dockerfile, and documentation.

2. **Risks:**
   - Windows CGO cross-compilation with MinGW + OpenVINO is unproven — may require native Windows runner as fallback
   - OpenVINO .pc files may not exist in the official archive — may need custom PKG_CONFIG_PATH
   - Archive bundling of .so/.dll files requires files to be at known paths inside the builder image

3. **Mitigations:**
   - If Windows CGO cross-compile fails, fall back to native Windows runner in GitHub Actions
   - If .pc files are missing, use CGO_CFLAGS/CGO_LDFLAGS env vars instead
   - Verify .so/.dll paths in the actual OpenVINO archive before finalizing

---

### Execution Handoff

After writing the plan, the user should choose an execution approach. If using subagent-driven development, dispatch one subagent per task with a two-stage review.
