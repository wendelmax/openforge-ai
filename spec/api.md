# API Specification

## Base URL

```
http://localhost:9090/v1
```

## Standard Headers

```http
Content-Type: application/json
Accept: application/json
X-Request-ID: <uuid>
```

## Standard Response Envelope

```json
{
  "object": "chat.completion",
  "id": "chatcmpl-xxx",
  "created": 1710000000,
  "model": "model-id",
  "choices": [...],
  "usage": {
    "prompt_tokens": 42,
    "completion_tokens": 128,
    "total_tokens": 170
  },
  "timing": {
    "ttft": "150ms",
    "total": "2.5s",
    "tokens_per_second": 51.2
  }
}
```

## Error Response

```json
{
  "error": {
    "code": "model_not_loaded",
    "message": "Model 'llama-3.2-3b' is not loaded. Use POST /model/load first.",
    "request_id": "abc-123"
  }
}
```

---

## Endpoints

### POST /v1/chat

Chat completion with conversation history.

**Request**:
```json
{
  "model": "llama-3.2-3b",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "What is Go?"}
  ],
  "temperature": 0.7,
  "top_p": 0.9,
  "max_tokens": 2048,
  "stream": false
}
```

**Response (non-streaming)**:
```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1710000000,
  "model": "llama-3.2-3b",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Go is a statically typed, compiled programming language..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 18,
    "completion_tokens": 42,
    "total_tokens": 60
  }
}
```

**Response (streaming — SSE)**:
```text
data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Go"},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":" is"},"finish_reason":null}]}

data: [DONE]
```

### POST /v1/completion

Text completion (non-chat).

**Request**:
```json
{
  "model": "codegemma-2b",
  "prompt": "package main\nimport \"fmt\"\nfunc main() {\n",
  "max_tokens": 128,
  "temperature": 0.2,
  "stop": ["\n\n"]
}
```

### POST /v1/embeddings

Generate embeddings for input text(s).

**Request**:
```json
{
  "model": "bge-small-en-v1.5",
  "input": ["Hello world", "Go programming"]
}
```

**Response**:
```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "index": 0,
      "embedding": [0.001, -0.023, ...]
    },
    {
      "object": "embedding",
      "index": 1,
      "embedding": [0.015, 0.003, ...]
    }
  ],
  "model": "bge-small-en-v1.5",
  "usage": {
    "prompt_tokens": 6,
    "total_tokens": 6
  }
}
```

### POST /v1/rerank

Rerank documents by relevance to query.

**Request**:
```json
{
  "model": "bge-reranker-v2-m3",
  "query": "What is Go?",
  "documents": [
    "Go is a programming language.",
    "Python is a programming language.",
    "The sky is blue."
  ],
  "top_n": 2
}
```

**Response**:
```json
{
  "object": "rerank",
  "model": "bge-reranker-v2-m3",
  "results": [
    {"index": 0, "score": 0.98, "document": "Go is a programming language."},
    {"index": 1, "score": 0.45, "document": "Python is a programming language."}
  ]
}
```

### POST /v1/model/load

Load a model into memory.

**Request**:
```json
{
  "model_id": "llama-3.2-3b",
  "device": "auto"
}
```

**Response**:
```json
{
  "object": "model",
  "id": "llama-3.2-3b",
  "status": "loaded",
  "device": "GPU",
  "timing": {
    "load_duration": "3.2s"
  }
}
```

### POST /v1/model/unload

Unload model from memory.

**Request**:
```json
{
  "model_id": "llama-3.2-3b"
}
```

**Response**:
```json
{
  "object": "model",
  "id": "llama-3.2-3b",
  "status": "unloaded"
}
```

### GET /v1/devices

List available compute devices.

**Response**:
```json
{
  "object": "list",
  "data": [
    {
      "id": "CPU",
      "name": "Intel Core i9-14900K",
      "type": "cpu",
      "available": true,
      "memory": 0
    },
    {
      "id": "GPU.0",
      "name": "Intel Arc A770",
      "type": "gpu",
      "available": true,
      "memory": 16777216
    },
    {
      "id": "NPU",
      "name": "Intel AI Boost",
      "type": "npu",
      "available": true,
      "memory": 8388608
    }
  ]
}
```

### GET /v1/models

List available models.

**Response**:
```json
{
  "object": "list",
  "data": [
    {
      "id": "llama-3.2-3b",
      "name": "Llama 3.2 3B",
      "path": "/models/llama-3.2-3b",
      "precision": "FP16",
      "loaded": true,
      "size": 2147483648
    },
    {
      "id": "bge-small-en-v1.5",
      "name": "BGE Small EN v1.5",
      "path": "/models/bge-small-en-v1.5",
      "precision": "FP32",
      "loaded": false,
      "size": 134217728
    }
  ]
}
```

### GET /v1/health

Health check.

**Response**:
```json
{
  "status": "ok",
  "version": "0.1.0",
  "uptime": "1h32m",
  "models_loaded": 2,
  "active_device": "GPU.0"
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| model_not_found | 404 | Model ID not found in registry |
| model_not_loaded | 503 | Model not loaded in memory |
| device_unavailable | 503 | Requested device not available |
| invalid_request | 400 | Malformed request body |
| inference_error | 500 | Error during inference |
| load_error | 500 | Error loading model |
| timeout | 504 | Inference exceeded timeout |
