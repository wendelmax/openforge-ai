# Stage 1: Build with OpenVINO
FROM ghcr.io/openforge-ai/builder:go1.26-ov2026.2.1 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN source build/builder-env.sh && \
    go build -ldflags="-s -w -X github.com/openforge-ai/openforge/cmd/openforge/cmd.Version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -o openforge ./cmd/openforge

# Stage 2: Minimal runtime with OpenVINO shared libs
FROM ubuntu:24.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -d /home/openforge -m openforge

# Copy OpenVINO runtime libraries
COPY --from=builder /opt/openvino/runtime/lib/intel64/libopenvino_c.so* /usr/local/lib/
COPY --from=builder /opt/openvino/runtime/lib/intel64/libopenvino.so* /usr/local/lib/
COPY --from=builder /opt/openvino/runtime/lib/intel64/libopenvino_intel_cpu_plugin.so /usr/local/lib/
COPY --from=builder /opt/openvino/runtime/lib/intel64/libopenvino_intel_gpu_plugin.so /usr/local/lib/
COPY --from=builder /opt/openvino/runtime/lib/intel64/libopenvino_intel_npu_plugin.so /usr/local/lib/
COPY --from=builder /opt/openvino/runtime/lib/intel64/libopenvino_ir_frontend.so* /usr/local/lib/
COPY --from=builder /opt/openvino/runtime/lib/intel64/cache.json /usr/local/lib/
RUN ldconfig

COPY --from=builder /build/openforge /usr/local/bin/openforge

USER openforge
WORKDIR /home/openforge

EXPOSE 9090
ENTRYPOINT ["openforge"]
CMD ["serve"]
