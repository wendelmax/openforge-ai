# Quickstart

Get OpenForge running in 5 minutes.

## Step 1: Install

```bash
curl -fsSL https://openforge.ai/install.sh | bash
```

## Step 2: Download a Model

```bash
openforge model pull llama-3.2-3b
```

This downloads Llama 3.2 3B in OpenVINO IR format (~2GB).

## Step 3: Start the Server

```bash
openforge serve
```

```
[INFO] server started  address=127.0.0.1:9090 device=GPU.0
[INFO] model loaded    model_id=llama-3.2-3b device=GPU.0
```

## Step 4: Chat

```bash
curl http://localhost:9090/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama-3.2-3b",
    "messages": [{"role": "user", "content": "What is Go?"}],
    "stream": false
  }'
```

Response:

```json
{
  "choices": [{"message": {"content": "Go is a statically typed..."}}],
  "usage": {"prompt_tokens": 5, "completion_tokens": 42}
}
```

## Step 5: Connect OpenCode

Configure OpenCode to use OpenForge as provider:

```json
{
  "provider": "openforge",
  "endpoint": "http://localhost:9090/v1",
  "model": "llama-3.2-3b"
}
```

## Step 6: Run a Skill

```bash
openforge skill run summarize --input doc.txt
```

## What's Next?

- Learn about [Architecture](../architecture.md)
- Browse the [Skills Catalog](../skills/catalog.md)
- Check [Benchmarks](../benchmarks.md)
- Read the [API Reference](../api-reference.md)
