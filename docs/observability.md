# Observability

OpenForge exposes metrics, logs, and traces for monitoring and debugging.

## Metrics

### Default Metrics (at `/metrics`)

| Metric | Type | Description |
|--------|:----:|-------------|
| `openforge_inferences_total` | Counter | Total inference requests |
| `openforge_inference_duration_ms` | Histogram | Inference latency (ms) |
| `openforge_ttft_ms` | Histogram | Time to first token (ms) |
| `openforge_tokens_per_second` | Gauge | Generated tokens per second |
| `openforge_tokens_total` | Counter | Total tokens generated |
| `openforge_cache_hits_total` | Counter | Cache hit count |
| `openforge_cache_misses_total` | Counter | Cache miss count |
| `openforge_models_loaded` | Gauge | Currently loaded models |
| `openforge_memory_usage_bytes` | Gauge | Process memory (RSS) |
| `openforge_device_load` | Gauge | Device utilization (0-100) |
| `openforge_errors_total` | Counter | Error count by type |
| `openforge_requests_active` | Gauge | In-flight requests |

### Labels

All metrics include labels:

| Label | Example | Description |
|-------|---------|-------------|
| `model` | `phi-3-mini` | Model identifier |
| `device` | `GPU.0` | Inference device |
| `endpoint` | `/v1/chat` | API endpoint |
| `status` | `200` | Response status |

## Logging

OpenForge uses structured JSON logging via `log/slog`.

### Example Log Output

```json
{"time":"2025-06-26T10:00:00Z","level":"INFO","msg":"inference completed","model":"phi-3-mini","device":"GPU.0","latency_ms":45,"tokens":128,"tokens_per_second":32.1}
{"time":"2025-06-26T10:00:01Z","level":"WARN","msg":"model not loaded","model_id":"mistral-7b"}
{"time":"2025-06-26T10:00:02Z","level":"ERROR","msg":"inference failed","model":"phi-3-mini","error":"device timeout","request_id":"abc-123"}
```

### Log Levels

| Level | Usage |
|:------|-------|
| `DEBUG` | Detailed debugging (tensor shapes, token IDs) |
| `INFO` | Normal operations (startup, inference, shutdown) |
| `WARN` | Degraded but continuing (fallback to CPU) |
| `ERROR` | Failed operations (model load failure, inference error) |

### Configuration

```yaml
logging:
  level: info           # debug, info, warn, error
  format: json          # json or text
  output: stdout        # stdout, stderr, or file path
```

## Tracing (OpenTelemetry)

Enable distributed tracing with OpenTelemetry:

```yaml
telemetry:
  enabled: true
  endpoint: "http://otel-collector:4318"
  service_name: "openforge"
  sampling_rate: 0.1    # Sample 10% of requests
```

### Traced Operations

| Span | Parent | Description |
|------|--------|-------------|
| `POST /v1/chat` | — | HTTP request |
| `engine.infer` | HTTP handler | Engine orchestration |
| `provider.infer` | engine.infer | Provider inference |
| `runtime.infer` | provider.infer | OpenVINO execution |
| `cache.lookup` | engine.infer | Cache check |
| `model.load` | — | Model loading |

## Health Check

```
GET /v1/health
```

```json
{
  "status": "ok",
  "version": "0.1.0",
  "commit": "abc1234",
  "uptime": "1h32m15s",
  "models_loaded": 2,
  "active_device": "GPU.0",
  "memory_mb": 2456,
  "cache_entries": 1284
}
```

## Debug Endpoints

| Endpoint | Description |
|:---------|-------------|
| `GET /debug/pprof/` | Go pprof profiles |
| `GET /debug/vars` | expvar metrics |
| `GET /metrics` | Prometheus metrics |

## Grafana Dashboard

A community-maintained Grafana dashboard is available:

```
https://grafana.com/grafana/dashboards/openforge
```

Includes panels for:
- Request rate and latency
- Device utilization
- Cache hit ratio
- Model memory usage
- Error rate by endpoint
- Tokens per second (real-time)
