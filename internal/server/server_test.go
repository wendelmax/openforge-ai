package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/internal/engine"
	"github.com/openforge-ai/openforge/runtime"
	"github.com/stretchr/testify/assert"
)

type testRuntime struct{}

func (t *testRuntime) ListDevices(ctx context.Context) ([]runtime.Device, error) {
	return []runtime.Device{
		{ID: "cpu", Name: "Intel CPU", Type: runtime.DeviceCPU, Available: true, Priority: 1},
		{ID: "gpu", Name: "Intel GPU", Type: runtime.DeviceGPU, Available: false, Priority: 2},
	}, nil
}

func (t *testRuntime) LoadModel(ctx context.Context, modelID, path, device string) error {
	return nil
}

func (t *testRuntime) UnloadModel(ctx context.Context, modelID string) error {
	return nil
}

func (t *testRuntime) ListModels(ctx context.Context) ([]runtime.ModelInfo, error) {
	return []runtime.ModelInfo{
		{ID: "phi-3-mini", Name: "Phi-3 Mini", Loaded: true},
		{ID: "bge-small", Name: "BGE Small EN", Loaded: false},
	}, nil
}

func (t *testRuntime) Infer(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (*runtime.InferenceResult, error) {
	return &runtime.InferenceResult{
		Text:   "mock response",
		Tokens: []int{1, 2, 3, 4, 5},
	}, nil
}

func (t *testRuntime) InferStream(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (<-chan string, error) {
	return nil, nil
}

func (t *testRuntime) Embed(ctx context.Context, modelID string, inputs []string, device string) (*runtime.EmbeddingResult, error) {
	n := len(inputs)
	embeddings := make([][]float32, n)
	for i := 0; i < n; i++ {
		embeddings[i] = []float32{1.0, 0.0, 0.0}
	}
	if len(inputs) > 1 {
		embeddings[1] = []float32{0.9, 0.1, 0.0}
	}
	return &runtime.EmbeddingResult{Embeddings: embeddings}, nil
}

func (t *testRuntime) DeviceForWorkload(workload, requestDevice string) string {
	if requestDevice != "" && requestDevice != "auto" {
		return requestDevice
	}
	return "cpu"
}

func (t *testRuntime) Benchmark(ctx context.Context, modelID string, iterations int, prompt string, maxTokens int) (runtime.BenchmarkResults, error) {
	return nil, nil
}

func (t *testRuntime) Close(ctx context.Context) error {
	return nil
}

func setupTestServer() *Server {
	gin.SetMode(gin.TestMode)
	rt := &testRuntime{}
	eng := engine.New(rt)
	cfg := config.Default()
	return New(eng, cfg)
}

func TestHealthEndpoint(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}

func TestV1HealthEndpoint(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthMethodNotAllowed(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRerankEndpoint(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"bge-small","query":"test query","documents":["doc1","doc2"]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/rerank", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp runtime.RerankResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "rerank", resp.Object)
	assert.Len(t, resp.Results, 2)
}

func TestRerankEndpointMissingModel(t *testing.T) {
	srv := setupTestServer()

	body := `{"query":"test query","documents":["doc1"]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/rerank", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRerankEndpointTopN(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"bge-small","query":"test query","documents":["doc1","doc2","doc3"],"top_n":1}`
	req := httptest.NewRequest(http.MethodPost, "/v1/rerank", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp runtime.RerankResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Results, 1)
}

func TestBenchmarkEndpoint(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"phi-3-mini","iterations":2}`
	req := httptest.NewRequest(http.MethodPost, "/v1/benchmark", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp runtime.BenchmarkResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "phi-3-mini", resp.Model)
}

func TestBenchmarkEndpointMissingModel(t *testing.T) {
	srv := setupTestServer()

	body := `{"iterations":2}`
	req := httptest.NewRequest(http.MethodPost, "/v1/benchmark", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatEndpoint(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"phi-3-mini","messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp runtime.ChatResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "chat.completion", resp.Object)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, "assistant", resp.Choices[0].Message.Role)
	assert.Equal(t, "mock response", resp.Choices[0].Message.Content)
}

func TestChatEndpointMissingModel(t *testing.T) {
	srv := setupTestServer()

	body := `{"messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatEndpointMissingMessages(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"phi-3-mini"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}



func TestCompletionEndpoint(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"phi-3-mini","prompt":"hello world"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/completion", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp runtime.CompletionResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "text_completion", resp.Object)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, "mock response", resp.Choices[0].Text)
}

func TestCompletionEndpointMissingModel(t *testing.T) {
	srv := setupTestServer()

	body := `{"prompt":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/completion", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompletionEndpointMissingPrompt(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"phi-3-mini"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/completion", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmbeddingsEndpoint(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"bge-small","input":["hello","world"]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp runtime.EmbeddingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "list", resp.Object)
	assert.Len(t, resp.Data, 2)
}

func TestEmbeddingsEndpointMissingModel(t *testing.T) {
	srv := setupTestServer()

	body := `{"input":["test"]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmbeddingsEndpointMissingInput(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"bge-small"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestModelLoadEndpoint(t *testing.T) {
	srv := setupTestServer()

	body := `{"model_id":"phi-3-mini","device":"cpu"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/model/load", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp runtime.ModelStatus
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "loaded", resp.Status)
}

func TestModelLoadEndpointMissingModelID(t *testing.T) {
	srv := setupTestServer()

	body := `{"device":"cpu"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/model/load", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestModelUnloadEndpoint(t *testing.T) {
	srv := setupTestServer()

	body := `{"model_id":"phi-3-mini"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/model/unload", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp runtime.ModelStatus
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unloaded", resp.Status)
}

func TestModelUnloadEndpointMissingModelID(t *testing.T) {
	srv := setupTestServer()

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/v1/model/unload", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListModelsEndpoint(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp runtime.ModelListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "list", resp.Object)
	assert.Len(t, resp.Data, 2)
}

func TestListDevicesEndpoint(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/v1/devices", nil)
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp runtime.DeviceListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "list", resp.Object)
	assert.Len(t, resp.Data, 2)
}

func TestRerankEndpointMissingQuery(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"bge-small","documents":["doc1"]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/rerank", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRerankEndpointMissingDocuments(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"bge-small","query":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/rerank", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatEndpointTemperature(t *testing.T) {
	srv := setupTestServer()

	body := `{"model":"phi-3-mini","messages":[{"role":"user","content":"hi"}],"temperature":0.5}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewUsage(t *testing.T) {
	u := NewUsage(10, 20)
	assert.Equal(t, 10, u.PromptTokens)
	assert.Equal(t, 20, u.CompletionTokens)
	assert.Equal(t, 30, u.TotalTokens)
}

func TestStrPtr(t *testing.T) {
	s := strPtr("hello")
	assert.NotNil(t, s)
	assert.Equal(t, "hello", *s)
}

func TestCosineSimilarity(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{1, 0, 0}
	assert.InDelta(t, float32(1.0), cosineSimilarity(a, b), 0.01)

	c := []float32{0, 1, 0}
	assert.InDelta(t, float32(0.0), cosineSimilarity(a, c), 0.01)

	d := []float32{2, 0, 0}
	assert.InDelta(t, float32(1.0), cosineSimilarity(a, d), 0.01)
}

func TestCosineSimilarityEmpty(t *testing.T) {
	assert.Equal(t, float32(0), cosineSimilarity(nil, []float32{1, 2}))
	assert.Equal(t, float32(0), cosineSimilarity([]float32{1, 2}, []float32{1, 2, 3}))
}

func TestRequestIDFromHeader(t *testing.T) {
	srv := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "custom-id")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, "custom-id", w.Header().Get("X-Request-ID"))
}

func TestRequestIDGenerated(t *testing.T) {
	srv := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	rid := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, rid)
}

func TestChatEndpointTopP(t *testing.T) {
	srv := setupTestServer()
	body := `{"model":"phi-3-mini","messages":[{"role":"user","content":"hi"}],"top_p":0.5}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChatEndpointTopK(t *testing.T) {
	srv := setupTestServer()
	body := `{"model":"phi-3-mini","messages":[{"role":"user","content":"hi"}],"top_k":50}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBenchmarkEndpointAutoLoad(t *testing.T) {
	srv := setupTestServer()
	body := `{"model":"bge-small"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/benchmark", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCompletionEndpointWithParams(t *testing.T) {
	srv := setupTestServer()
	body := `{"model":"phi-3-mini","prompt":"hello","temperature":0.8,"top_p":0.95}`
	req := httptest.NewRequest(http.MethodPost, "/v1/completion", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInvalidJSON(t *testing.T) {
	srv := setupTestServer()
	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestModelLoadNoDevice(t *testing.T) {
	srv := setupTestServer()
	srv.config.Models.Device = "cpu"
	body := `{"model_id":"phi-3-mini"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/model/load", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBenchmarkEndpointDefaults(t *testing.T) {
	srv := setupTestServer()
	body := `{"model":"phi-3-mini"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/benchmark", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

type errorRuntime struct {
	testRuntime
	listModelsFail    bool
	listDevicesFail   bool
	unloadModelFail   bool
	inferFail         bool
	inferStreamFail   bool
	embedFail         bool
}

func (e *errorRuntime) InferStream(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (<-chan string, error) {
	if e.inferStreamFail {
		return nil, assert.AnError
	}
	ch := make(chan string)
	close(ch)
	return ch, nil
}

func (e *errorRuntime) Embed(ctx context.Context, modelID string, inputs []string, device string) (*runtime.EmbeddingResult, error) {
	if e.embedFail {
		return nil, assert.AnError
	}
	return e.testRuntime.Embed(ctx, modelID, inputs, device)
}

func (e *errorRuntime) ListModels(ctx context.Context) ([]runtime.ModelInfo, error) {
	if e.listModelsFail {
		return nil, assert.AnError
	}
	return e.testRuntime.ListModels(ctx)
}

func (e *errorRuntime) ListDevices(ctx context.Context) ([]runtime.Device, error) {
	if e.listDevicesFail {
		return nil, assert.AnError
	}
	return e.testRuntime.ListDevices(ctx)
}

func (e *errorRuntime) UnloadModel(ctx context.Context, modelID string) error {
	if e.unloadModelFail {
		return assert.AnError
	}
	return e.testRuntime.UnloadModel(ctx, modelID)
}

func (e *errorRuntime) Infer(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (*runtime.InferenceResult, error) {
	if e.inferFail {
		return nil, assert.AnError
	}
	return e.testRuntime.Infer(ctx, modelID, prompt, params)
}

func setupErrorServer(failures ...func(*errorRuntime)) *Server {
	gin.SetMode(gin.TestMode)
	rt := &errorRuntime{}
	for _, f := range failures {
		f(rt)
	}
	eng := engine.New(rt)
	cfg := config.Default()
	return New(eng, cfg)
}

func TestListModelsError(t *testing.T) {
	srv := setupErrorServer(func(rt *errorRuntime) { rt.listModelsFail = true })
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListDevicesError(t *testing.T) {
	srv := setupErrorServer(func(rt *errorRuntime) { rt.listDevicesFail = true })
	req := httptest.NewRequest(http.MethodGet, "/v1/devices", nil)
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestModelUnloadError(t *testing.T) {
	srv := setupErrorServer(func(rt *errorRuntime) { rt.unloadModelFail = true })
	body := `{"model_id":"phi-3-mini"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/model/unload", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestChatInferenceError(t *testing.T) {
	srv := setupErrorServer(func(rt *errorRuntime) { rt.inferFail = true })
	body := `{"model":"phi-3-mini","messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCompletionInferenceError(t *testing.T) {
	srv := setupErrorServer(func(rt *errorRuntime) { rt.inferFail = true })
	body := `{"model":"phi-3-mini","prompt":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/completion", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEmbeddingsInferenceError(t *testing.T) {
	srv := setupErrorServer(func(rt *errorRuntime) { rt.embedFail = true })
	body := `{"model":"bge-small","input":["test"]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestBenchmarkEndpointUnloadedModel(t *testing.T) {
	srv := setupTestServer()
	body := `{"model":"bge-small"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/benchmark", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChatStreamError(t *testing.T) {
	srv := setupErrorServer(func(rt *errorRuntime) { rt.inferStreamFail = true })
	body := `{"model":"phi-3-mini","messages":[{"role":"user","content":"hello"}],"stream":true}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestChatStreamSuccess(t *testing.T) {
	srv := setupErrorServer()
	body := `{"model":"phi-3-mini","messages":[{"role":"user","content":"hello"}],"stream":true}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bodyReader(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.gin.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRenderJSON(t *testing.T) {
	srv := setupTestServer()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	srv.renderJSON(c, map[string]string{"key": "value"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func bodyReader(t *testing.T, body string) *strings.Reader {
	t.Helper()
	return strings.NewReader(body)
}
