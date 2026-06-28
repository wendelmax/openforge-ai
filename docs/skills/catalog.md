# Skills Catalog

Skills are reusable AI pipelines that solve specific development tasks.

## Language Skills

### Go
Generate, review, test, and refactor Go code.

```bash
openforge skill run go:generate --param description="REST API with Gin"
openforge skill run go:test --param file=handler.go
openforge skill run go:review --param file=main.go
openforge skill run go:refactor --param file=legacy.go --param pattern="modernize"
```

**Commands:** `generate`, `test`, `review`, `refactor`, `doc`, `benchmark`

---

### Java / Spring
Generate Spring Boot projects, services, controllers, and tests.

```bash
openforge skill run java:spring-init --param name=my-app --param deps=web,jpa,postgres
openforge skill run java:generate-service --param entity=User
openforge skill run java:test --param class=UserService
```

**Commands:** `spring-init`, `generate-service`, `generate-controller`, `test`, `review`

---

### Python
Generate, test, and review Python code with focus on FastAPI, Django, and data science.

```bash
openforge skill run python:fastapi-init --param name=api
openforge skill run python:generate-model --param fields="id,name,email"
openforge skill run python:test --param file=services.py
```

**Commands:** `fastapi-init`, `generate-model`, `test`, `review`, `migrate`

---

### TypeScript / JavaScript
Generate Node.js, NestJS, and React code.

```bash
openforge skill run ts:nest-init --param name=backend
openforge skill run ts:generate-module --param name=users
openforge skill run ts:react-component --param name=UserProfile
```

**Commands:** `nest-init`, `generate-module`, `react-component`, `test`, `review`

---

### Rust
Generate Rust code with focus on safety and performance.

```bash
openforge skill run rust:init --param name=cli-tool
openforge skill run rust:generate-struct --param fields="name:String,age:u32"
```

**Commands:** `init`, `generate-struct`, `test`, `review`

---

## Framework Skills

### Spring Boot
Build production-ready Spring Boot applications.

```bash
openforge skill run spring:controller --param entity=Product
openforge skill run spring:repository --param entity=Product
openforge skill run spring:service --param entity=Product
```

### Docker
Generate Dockerfiles and docker-compose configurations.

```bash
openforge skill run docker:generate --param language=go --param port=8080
openforge skill run docker:compose --param services="api,db,redis"
```

### Kubernetes
Generate Kubernetes manifests.

```bash
openforge skill run k8s:deployment --param name=api --param image=myapp:latest
openforge skill run k8s:service --param name=api --param port=80
```

---

## Infrastructure Skills

### Git / GitHub
Generate .gitignore, CI workflows, and PR templates.

```bash
openforge skill run git:init --param language=go
openforge skill run github:ci --param language=go --param test=true
```

### SQL
Design and generate database schemas.

```bash
openforge skill run sql:generate --param tables="users,posts,comments"
openforge skill run sql:migrate --param description="add user email"
```

### REST API
Design and document REST APIs.

```bash
openforge skill run rest:endpoint --param method=GET --param path=/users
openforge skill run rest:openapi --param file=routes.go
```

---

## Quality Skills

### Testing
Generate unit tests, integration tests, and mocks.

```bash
openforge skill run test:generate --param file=calculator.go
openforge skill run test:integration --param service=UserService
```

### Code Review
Review code for bugs, security issues, and best practices.

```bash
openforge skill run review:security --param file=auth.go
openforge skill run review:performance --param file=database.go
```

### Documentation
Generate GoDoc, README, and API docs from code.

```bash
openforge skill run doc:readme --param file=main.go
openforge skill run doc:godoc --param package=./internal/...
```

---

## OpenForge Skills

### Benchmark
Benchmark models and devices.

```bash
openforge skill run of:benchmark --param model=phi-3-mini
```

### Convert Model
Convert HuggingFace models to OpenVINO IR.

```bash
openforge skill run of:convert --param model=microsoft/phi-3-mini
```

### Export Metrics
Export telemetry data.

```bash
openforge skill run of:metrics --param format=prometheus
```

---

## Creating Custom Skills

Skills are YAML files. See [Creating Skills](creating-skills.md).

```bash
openforge skill create my-skill
openforge skill run my-skill --param input=data.txt
```
