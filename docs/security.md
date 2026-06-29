# Security

## Philosophy

OpenForge is designed for **offline-first** operation. All inference runs locally, on your hardware, with your data. No data leaves your machine unless you explicitly enable telemetry.

## Data Flow

```mermaid
graph LR
    UI["User Input"] --> OF["OpenForge Process"] --> OV["OpenVINO Runtime"] --> HW["Local Hardware"]
    UI -.- B["All data stays here"]
    HW -.- B
```

- **No external API calls** during inference
- **No data upload** to cloud services
- **No telemetry** by default (opt-in only)
- **No analytics** embedded in the binary

## Permissions

### File System

OpenForge accesses:
- Model directory (read-only after loading)
- Cache directory (read/write, configurable)
- Config file (read-only)
- Plugin directory (read, if configured)

No access to:
- System files outside configured directories
- Environment variables (except `OPENFORGE_*`)
- Network sockets (except configured server port)

### Network

- **Server**: listens on `127.0.0.1:9090` by default (localhost only)
- **No outbound connections** in normal operation
- Telemetry (if enabled) uses OpenTelemetry to a configurable collector

### Process

- Runs as the invoking user (no privilege escalation)
- No subprocess execution
- No dynamic code loading (plugin loading restricted to `.so` files)

## Sandboxing Recommendations

For multi-tenant or high-security environments:

```bash
# Run as dedicated user
sudo useradd -r -s /bin/false openforge
sudo -u openforge openforge serve

# Use Linux namespaces
unshare -r -n openforge serve --port 9090

# Use AppArmor/SELinux
# (profile provided in hack/apparmor/)
```

## Plugin Security

Plugins (`.so` files) run in the same process as OpenForge.

**Risks:**
- Plugins have full access to the process memory
- Plugins can execute arbitrary code

**Mitigation:**
- Only load plugins from trusted sources
- Verify plugin hash before loading
- Review plugin source code when possible
- Run with dedicated user (see above)

## Telemetry

Telemetry is **disabled by default** and **opt-in**.

When enabled, the following data is collected:
- Model name and version
- Inference latency (min/max/avg)
- Device type
- Cache hit rate
- Error counts

**Never collected:**
- Prompt content
- Generated text
- File contents
- User identity
- Network metadata

## Reporting Vulnerabilities

Report security vulnerabilities to **security@openforge.ai**.

We will:
1. Acknowledge receipt within 24 hours
2. Investigate and fix within 7 days
3. Release a patch with CVE details
4. Credit the reporter (unless anonymity is requested)

## Known Security Considerations

| Concern | Status | Notes |
|---------|:------:|-------|
| GPU memory isolation | ⚠️ | GPU memory is shared with other processes |
| NPU memory isolation | ⚠️ | NPU memory is shared with other processes |
| Inference timing attacks | ⚠️ | Timing may leak input length |
| Model file tampering | ⚠️ | Verify model checksums |
| Plugin sandboxing | 🔧 | Planned for v1.0 |
