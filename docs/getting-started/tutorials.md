# Tutorials

## Beginner

### 1. Your First Chat

```bash
openforge serve --model llama-3.2-3b

curl -X POST http://localhost:9090/v1/chat \
  -d '{"model":"llama-3.2-3b","messages":[{"role":"user","content":"Write a hello world in Go"}]}'
```

### 2. Generate Embeddings

```python
import requests

resp = requests.post("http://localhost:9090/v1/embeddings", json={
    "model": "bge-small-en-v1.5",
    "input": ["Go is compiled", "Python is interpreted"]
})
print(resp.json()["data"][0]["embedding"][:5])
# [0.023, -0.045, 0.012, ...]
```

### 3. Semantic Search with Rerank

```bash
curl -X POST http://localhost:9090/v1/rerank \
  -d '{
    "model": "bge-reranker-v2-m3",
    "query": "concurrency patterns",
    "documents": [
      "Go uses goroutines for concurrency",
      "Java uses threads",
      "JavaScript is single-threaded"
    ],
    "top_n": 2
  }'
```

## Intermediate

### 4. Build a RAG Pipeline

```python
# 1. Ingest documents
chunks = chunk_document("docs.txt")
for chunk in chunks:
    emb = openforge.embed(chunk)
    vector_db.store(chunk, emb)

# 2. Query
query = "How do I install OpenForge?"
results = vector_db.search(openforge.embed(query))
context = "\n".join(results)

# 3. Generate with context
response = openforge.chat(
    f"Context: {context}\n\nQuestion: {query}"
)
```

### 5. Create and Run a Skill

```yaml
# skills/code-review.yaml
name: code-review
description: Review Go code for common issues
steps:
  - type: prompt
    model: llama-3.2-3b
    system: "Review this Go code:"
    input: "{{code}}"
    output: review

  - type: prompt
    model: llama-3.2-3b
    system: "Suggest fixes for: {{review}}"
    output: fixes
```

```bash
openforge skill run code-review --param code="$(cat main.go)"
```

### 6. Benchmark Your Hardware

```bash
openforge benchmark --model llama-3.2-3b --device auto
```

Output:
```
Model: llama-3.2-3b
Device: GPU.0 (Intel Arc A770)
Latency: 45ms/token
Throughput: 22.2 tok/s
Memory: 3.2 GB
```

## Advanced

### 7. Custom Plugin

```go
// my-plugin/main.go
package main

import "github.com/openforge-ai/openforge/runtime"

type MyProvider struct {}

func (p *MyProvider) Name() string { return "my-provider" }
func (p *MyProvider) Infer(req runtime.InferenceRequest) runtime.InferenceResponse {
    // custom logic
}
```

### 8. Multi-Model Pipeline

```yaml
# skills/analyze.yaml
steps:
  - type: embed
    model: bge-small-en-v1.5
    input: "{{document}}"
    output: embedding

  - type: prompt
    model: llama-3.2-3b
    system: "Summarize: {{document}}"
    output: summary

  - type: prompt
    model: phi-3-mini
    system: "Extract key points from: {{summary}}"
    output: keypoints
```
