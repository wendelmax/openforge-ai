# Showcase

Real projects and workflows powered by OpenForge.

## VS Code + OpenCode Integration

Developers using OpenForge as an OpenCode provider get:

```
┌─────────────────────────────────────────────┐
│  VS Code                                     │
│  ┌───────────────────────────────────────┐   │
│  │  func main() {                        │   │
│  │    fmt.Println("Hello, █")            │   │
│  │  }                                    │   │
│  └───────────────────────────────────────┘   │
│  ┌───────────────────────────────────────┐   │
│  │  OpenForge: World!                     │   │
│  │  [Tab] to accept  [Esc] to dismiss    │   │
│  └───────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

**Autocomplete latency**: ~150ms on NPU

---

## CLI Code Generation

```bash
# Spring Boot project in seconds
openforge skill run java:spring-init \
  --param name=order-service \
  --param deps=web,jpa,postgres,security

# Generates:
#   order-service/
#   ├── pom.xml
#   ├── src/main/java/com/example/
#   │   ├── OrderServiceApplication.java
#   │   ├── controller/OrderController.java
#   │   ├── service/OrderService.java
#   │   ├── repository/OrderRepository.java
#   │   └── model/Order.java
#   ├── src/main/resources/
#   │   └── application.yml
#   └── src/test/java/com/example/
#       └── OrderServiceApplicationTests.java
```

---

## RAG Pipeline for Documentation

```
Developer query: "How do I configure OpenForge for NPU?"
        │
        ▼
[Embedding: BGE Small → search vector DB]
        │
        ▼
[Found 3 relevant doc chunks]
        │
        ▼
[Rerank: BGE Reranker → top 2]
        │
        ▼
[Generate: Llama 3.2 3B → answer with citations]
        │
        ▼
"To use NPU, set `device: NPU` in config.yaml.
 The NPU is available on Intel Core Ultra processors.
 Source: docs/getting-started/installation.md (line 42)"
```

**End-to-end latency**: ~800ms

---

## Go Microservice Generator

```bash
openforge skill run go:generate \
  --param description="REST API for user management with PostgreSQL"

# Creates
#   api/
#   ├── handler/user.go
#   ├── middleware/auth.go
#   ├── model/user.go
#   ├── repository/user.go
#   ├── service/user.go
#   ├── router.go
#   ├── main.go
#   ├── Dockerfile
#   ├── docker-compose.yml
#   └── README.md
```

---

## Code Review Pipeline

```bash
# Review all changed files in a PR
git diff --name-only main..feature | \
  xargs -I {} openforge skill run review:security --param file={}
```

Output:
```
Review of auth.go:
───────────────────
🔴 CRITICAL: JWT secret hardcoded (line 15)
🟡 WARNING: No rate limiting on /login (line 42)
🟢 INFO: Password hashing uses bcrypt (good)
🟢 INFO: SQL injection prevention via params (good)
```

---

## Containerized Development

```yaml
# docker-compose.yml
services:
  openforge:
    image: openforge/openforge:latest
    ports:
      - "9090:9090"
    volumes:
      - ./models:/models
      - openforge-data:/data
    devices:
      - /dev/dri:/dev/dri  # GPU passthrough
    environment:
      - OPENFORGE_MODELS_DEVICE=GPU.0

  app:
    build: .
    depends_on:
      - openforge
    environment:
      - AI_ENDPOINT=http://openforge:9090/v1
```

---

## Your Project Here

OpenForge is built for the community. [Share your use case](https://github.com/openforge-ai/openforge/discussions) and we'll feature it here.

**Current community projects:**
- CLI-based Spring Boot generator
- VS Code extension for Go developers
- Automated PR reviewer for GitHub Actions
- Documentation RAG chatbot
- Local-first code assistant for air-gapped environments
