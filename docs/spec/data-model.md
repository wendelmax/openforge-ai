# Data Model

## Model

```go
type Model struct {
    ID        string            `json:"id"`
    Name      string            `json:"name"`
    Path      string            `json:"path"`
    Device    string            `json:"device"`
    Precision string            `json:"precision"`
    Loaded    bool              `json:"loaded"`
    LoadedAt  time.Time         `json:"loaded_at,omitempty"`
    Size      int64             `json:"size"`
    Version   string            `json:"version"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}
```

## InferenceRequest

```go
type InferenceRequest struct {
    ModelID    string            `json:"model_id"`
    Prompt     string            `json:"prompt"`
    System     string            `json:"system,omitempty"`
    Messages   []Message         `json:"messages,omitempty"`
    Temperature float32          `json:"temperature,omitempty"`
    TopK       int               `json:"top_k,omitempty"`
    TopP       float32           `json:"top_p,omitempty"`
    MaxTokens  int               `json:"max_tokens,omitempty"`
    Context    []int             `json:"context,omitempty"`
    Device      string           `json:"device,omitempty"`
    Stream     bool              `json:"stream,omitempty"`
}
```

## InferenceResponse

```go
type InferenceResponse struct {
    ModelID   string     `json:"model_id"`
    Text      string     `json:"text"`
    Tokens    []int      `json:"tokens,omitempty"`
    LogProbs  []float32  `json:"logprobs,omitempty"`
    Usage     Usage      `json:"usage"`
    Timing    Timing     `json:"timing"`
}
```

## Message

```go
type Message struct {
    Role    string `json:"role"`    // "system", "user", "assistant"
    Content string `json:"content"`
}
```

## EmbeddingRequest

```go
type EmbeddingRequest struct {
    ModelID string   `json:"model_id"`
    Input   []string `json:"input"`
    Device  string   `json:"device,omitempty"`
}
```

## EmbeddingResponse

```go
type EmbeddingResponse struct {
    ModelID    string     `json:"model_id"`
    Embeddings [][]float32 `json:"embeddings"`
    Usage      Usage      `json:"usage"`
    Timing     Timing     `json:"timing"`
}
```

## RerankRequest

```go
type RerankRequest struct {
    ModelID    string   `json:"model_id"`
    Query      string   `json:"query"`
    Documents  []string `json:"documents"`
    TopN       int      `json:"top_n,omitempty"`
}
```

## RerankResponse

```go
type RerankResult struct {
    Index     int     `json:"index"`
    Score     float32 `json:"score"`
    Document  string  `json:"document"`
}

type RerankResponse struct {
    ModelID  string         `json:"model_id"`
    Results  []RerankResult `json:"results"`
    Usage    Usage          `json:"usage"`
    Timing   Timing         `json:"timing"`
}
```

## Usage

```go
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

## Timing

```go
type Timing struct {
    TTFT       time.Duration `json:"ttft"`        // time to first token
    Total      time.Duration `json:"total"`
    TokensPerSecond float64  `json:"tokens_per_second"`
}
```

## Session

```go
type Session struct {
    ID        string    `json:"id"`
    ModelID   string    `json:"model_id"`
    Messages  []Message `json:"messages"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Context   []int     `json:"context,omitempty"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}
```

## Device

```go
type Device struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Type      string `json:"type"`   // "cpu", "gpu", "npu"
    Available bool   `json:"available"`
    Memory    int64  `json:"memory,omitempty"`
    Priority  int    `json:"priority"`
}
```

## Config

```go
type Config struct {
    Server    ServerConfig    `mapstructure:"server"`
    Models    ModelsConfig    `mapstructure:"models"`
    Cache     CacheConfig     `mapstructure:"cache"`
    Logging   LoggingConfig   `mapstructure:"logging"`
    Telemetry TelemetryConfig `mapstructure:"telemetry"`
}

type ServerConfig struct {
    Host    string `mapstructure:"host"`
    Port    int    `mapstructure:"port"`
    Timeout int    `mapstructure:"timeout"`
}

type ModelsConfig struct {
    Path     string `mapstructure:"path"`
    Default  string `mapstructure:"default"`
    Device   string `mapstructure:"device"`
}

type CacheConfig struct {
    Enabled bool   `mapstructure:"enabled"`
    Backend string `mapstructure:"backend"`
    Path    string `mapstructure:"path"`
    MaxSize int64  `mapstructure:"max_size"`
}

type LoggingConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"`
    Output string `mapstructure:"output"`
}

type TelemetryConfig struct {
    Enabled bool `mapstructure:"enabled"`
}
```

## Cache (SQLite)

```sql
CREATE TABLE embeddings_cache (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    model_id   TEXT NOT NULL,
    input_hash TEXT NOT NULL,
    embedding  BLOB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(model_id, input_hash)
);

CREATE TABLE response_cache (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    model_id    TEXT NOT NULL,
    input_hash  TEXT NOT NULL,
    response    TEXT NOT NULL,
    parameters  TEXT NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(model_id, input_hash, parameters)
);

CREATE TABLE sessions (
    id          TEXT PRIMARY KEY,
    model_id    TEXT NOT NULL,
    messages    TEXT NOT NULL,
    context     BLOB,
    metadata    TEXT,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
