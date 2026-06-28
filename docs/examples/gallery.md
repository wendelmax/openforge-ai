# Examples Gallery

Complete, runnable examples organized by use case.

## Example 1: CRUD API in Java with Spring Boot

```bash
openforge skill run java:spring-init \
  --param name=user-service \
  --param deps=web,jpa,postgres,validation

cd user-service

# User entity
openforge skill run java:generate-entity \
  --param name=User \
  --param fields="id:Long,name:String,email:String,createdAt:LocalDateTime"

# REST Controller
openforge skill run java:generate-controller \
  --param entity=User \
  --param endpoints="GET,POST,PUT,DELETE"

# Run
./mvnw spring-boot:run
```

**Generated**: `User.java`, `UserController.java`, `UserService.java`, `UserRepository.java`, `application.yml`

---

## Example 2: Generate Unit Tests

```bash
# For a Go package
openforge skill run go:test \
  --param file=service.go \
  --param coverage=90

# For a Java class
openforge skill run java:test \
  --param class=UserService \
  --param framework=junit5
```

**Output**: Complete test files with mocks, edge cases, and table-driven tests.

---

## Example 3: Refactor Legacy Code

```bash
openforge skill run go:refactor \
  --param file=legacy.go \
  --param pattern="modernize"

# Detects:
# - Global variables → dependency injection
# - Missing error handling → proper propagation
# - Old syntax → modern Go idioms
# - Missing tests → test suggestions
```

---

## Example 4: Convert HuggingFace Model to OpenVINO

```bash
# Using the convert skill
openforge skill run of:convert \
  --param model=microsoft/phi-3-mini-4k-instruct \
  --param precision=int4

# Files created:
#   phi-3-mini-int4/
#   ├── openvino_model.xml
#   ├── openvino_model.bin
#   ├── tokenizer.json
#   └── config.json

# Benchmark it
openforge skill run of:benchmark \
  --param model=phi-3-mini-int4 \
  --param device=auto
```

---

## Example 5: Hardware Benchmark

```bash
# Full benchmark report
openforge benchmark --all --output benchmark.json

# Sample output
{
  "model": "phi-3-mini-int4",
  "device": "GPU.0",
  "token_per_second": 32.1,
  "ttft_ms": 45.2,
  "memory_mb": 2400,
  "latency_p50_ms": 31.2,
  "latency_p95_ms": 48.7,
  "latency_p99_ms": 62.3
}
```

---

## Example 6: Chat with Context

```python
import requests

API = "http://localhost:9090/v1"

# Create a session with context
session = requests.post(f"{API}/chat", json={
    "model": "llama-3.2-3b",
    "messages": [
        {"role": "system", "content": "You are a Go expert."},
        {"role": "user", "content": "Explain goroutines"}
    ]
}).json()

# Continue the conversation
response = requests.post(f"{API}/chat", json={
    "model": "llama-3.2-3b",
    "messages": [
        {"role": "system", "content": "You are a Go expert."},
        {"role": "user", "content": "Explain goroutines"},
        {"role": "assistant", "content": session["choices"][0]["message"]["content"]},
        {"role": "user", "content": "Show me an example"}
    ]
}).json()
```

---

## Example 7: RAG Pipeline

```python
import requests, json

API = "http://localhost:9090/v1"

# 1. Embed a query
query = "How do I use NPU with OpenForge?"
embed = requests.post(f"{API}/embeddings", json={
    "model": "bge-small-en-v1.5",
    "input": [query]
}).json()

query_vector = embed["data"][0]["embedding"]

# 2. Search vector DB (pseudocode)
results = vector_db.search(query_vector, top_k=5)

# 3. Rerank results
rerank = requests.post(f"{API}/rerank", json={
    "model": "bge-reranker-v2-m3",
    "query": query,
    "documents": [r["text"] for r in results],
    "top_n": 3
}).json()

# 4. Generate answer with context
context = "\n".join([r["document"] for r in rerank["results"]])
chat = requests.post(f"{API}/chat", json={
    "model": "llama-3.2-3b",
    "messages": [
        {"role": "system", "content": f"Context:\n{context}"},
        {"role": "user", "content": query}
    ]
}).json()

print(chat["choices"][0]["message"]["content"])
```

---

## Example 8: CI/CD Integration (GitHub Actions)

```yaml
# .github/workflows/ai-code-review.yml
name: AI Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    services:
      openforge:
        image: openforge/openforge:latest
        ports:
          - 9090:9090
        volumes:
          - /opt/models:/models

    steps:
      - uses: actions/checkout@v4

      - name: Run AI Code Review
        run: |
          git diff origin/main --name-only | \
            xargs -I {} openforge skill run review:security \
              --param file={} \
              --endpoint http://localhost:9090/v1

      - name: Post Review Comments
        uses: openforge/review-action@v1
```

---

## Example 9: Custom VS Code Extension

```typescript
// extension.ts
import * as vscode from 'vscode';

export function activate(context: vscode.ExtensionContext) {
    const provider = vscode.languages.registerInlineCompletionItemProvider(
        { pattern: '**' },
        {
            async provideInlineCompletionItems(document, position) {
                const text = document.getText();
                const response = await fetch('http://localhost:9090/v1/completion', {
                    method: 'POST',
                    body: JSON.stringify({
                        model: 'codegemma-2b',
                        prompt: text,
                        max_tokens: 32
                    })
                });
                const data = await response.json();
                return [new vscode.InlineCompletionItem(data.text)];
            }
        }
    );
    context.subscriptions.push(provider);
}
```

---

## Example 10: Multi-Step Skill Pipeline

```yaml
# skills/document-review.yaml
name: document-review
description: Review and improve technical documentation
version: 1.0.0

inputs:
  document:
    type: string
    required: true
  language:
    type: string
    default: en

steps:
  - id: grammar
    type: prompt
    model: phi-3-mini
    system: "Fix grammar and spelling in this {{.inputs.language}} text. Return only corrections."
    user: "{{.inputs.document}}"
    output: grammar_fixes

  - id: clarity
    type: prompt
    model: llama-3.2-3b
    system: "Improve clarity of this technical document. Suggest specific rewrites."
    user: "{{.steps.grammar.output}}"
    output: clarity_suggestions

  - id: format
    type: format
    template: |
      # Document Review

      ## Grammar & Spelling
      {{.steps.grammar.output}}

      ## Clarity Improvements
      {{.steps.clarity.output}}
    output: result
```

```bash
openforge skill run document-review \
  --param document="$(cat README.md)"
```
