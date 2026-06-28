# Creating Skills

Skills are YAML pipelines that turn prompts into reusable, composable tools.

## Anatomy of a Skill

```yaml
name: code-review
description: Review Go code for bugs and best practices
version: 1.0.0
author: OpenForge

inputs:
  code:
    type: string
    description: Source code to review
    required: true

  language:
    type: string
    description: Programming language
    default: go

steps:
  - id: review
    type: prompt
    model: llama-3.2-3b
    system: |
      You are an expert {{.inputs.language}} reviewer.
      Analyze the code for:
      - Bugs and logic errors
      - Security vulnerabilities
      - Performance issues
      - Best practices violations
    user: "{{.inputs.code}}"
    output: review_result

  - id: format
    type: format
    template: |
      ## Review Results

      {{.steps.review.output}}
    output: result
```

## Step Types

### `prompt`
Execute an LLM prompt.

```yaml
- id: step1
  type: prompt
  model: llama-3.2-3b          # Model to use
  system: "You are a..."       # System prompt (optional)
  user: "{{input}}"            # User message (supports templates)
  temperature: 0.7             # Optional (default: 0.7)
  max_tokens: 2048             # Optional (default: 2048)
  output: my_output            # Variable to store result
```

### `embed`
Generate embeddings for text.

```yaml
- id: step1
  type: embed
  model: bge-small-en-v1.5
  input: "{{text}}"
  output: embedding
```

### `rerank`
Rerank documents by relevance.

```yaml
- id: step1
  type: rerank
  model: bge-reranker-v2-m3
  query: "{{query}}"
  documents: "{{documents}}"
  top_n: 5
  output: ranked
```

### `format`
Format text using Go templates.

```yaml
- id: step1
  type: format
  template: |
    # {{.inputs.title}}
    {{.steps.content.output}}
  output: result
```

### `condition`
Conditional branching.

```yaml
- id: check
  type: condition
  if: '{{contains .steps.review.output "CRITICAL"}}'
  then:
    - type: prompt
      model: llama-3.2-3b
      system: "Explain why this is critical:"
      input: "{{.steps.review.output}}"
      output: explanation
  else:
    - type: format
      template: "No critical issues found."
      output: explanation
```

### `loop`
Iterate over items.

```yaml
- id: process
  type: loop
  over: "{{.inputs.files}}"
  as: file
  steps:
    - type: prompt
      model: llama-3.2-3b
      system: "Review this file: {{.item.file}}"
      user: "{{.item.content}}"
      output: "review_{{.index}}"
```

## Templates

Templates use Go template syntax with these variables:

| Variable | Description |
|----------|-------------|
| `{{.inputs.name}}` | Skill input parameter |
| `{{.steps.step_id.output}}` | Output from a previous step |
| `{{.env.VAR_NAME}}` | Environment variable |

Functions available:

| Function | Description |
|----------|-------------|
| `{{lower "TEXT"}}` | Lowercase |
| `{{upper "text"}}` | Uppercase |
| `{{trim s}}` | Trim whitespace |
| `{{contains s substr}}` | String contains |
| `{{replace s old new}}` | String replace |
| `{{split s sep}}` | String split |
| `{{join arr sep}}` | Array join |
| `{{len arr}}` | Array length |
| `{{slice arr start end}}` | Array slice |
| `{{json .obj}}` | JSON encode |
| `{{now}}` | Current timestamp |

## Testing Skills

```bash
# Dry run (prints steps without execution)
openforge skill test my-skill --param code="func main() {}"

# Run with verbose output
openforge skill run my-skill --param code="func main() {}" --verbose

# Validate YAML syntax
openforge skill validate my-skill.yaml
```

## Publishing Skills

Skills can be shared as files or hosted in repositories:

```bash
# Local
openforge skill install ~/skills/my-skill.yaml

# Remote
openforge skill install https://skills.openforge.ai/go/review.yaml

# Marketplace (future)
openforge skill search code-review
openforge skill publish my-skill.yaml
```

## Best Practices

1. **One responsibility per skill** — do one thing well
2. **Use descriptive IDs** — step IDs become variable names
3. **Set appropriate models** — use small models for simple tasks
4. **Handle errors** — use `condition` steps for edge cases
5. **Test with real code** — verify output quality
6. **Version your skills** — bump version on breaking changes
7. **Document inputs** — clear descriptions for each parameter
8. **Keep it simple** — prefer fewer steps over clever pipelines
