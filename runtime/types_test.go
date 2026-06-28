package runtime

import (
	"encoding/json"
	"testing"
)

func TestChatRequest_WithDevice(t *testing.T) {
	req := ChatRequest{
		Model:  "phi-3-mini",
		Device: "NPU",
		Messages: []Message{{Role: "user", Content: "hi"}},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ChatRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Device != "NPU" {
		t.Errorf("expected NPU, got %q", decoded.Device)
	}
}

func TestChatRequest_WithoutDevice(t *testing.T) {
	req := ChatRequest{
		Model:  "phi-3-mini",
		Messages: []Message{{Role: "user", Content: "hi"}},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"model":"phi-3-mini","messages":[{"role":"user","content":"hi"}]}` {
		t.Errorf("unexpected JSON with device omitted: %s", string(data))
	}
	var decoded ChatRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Device != "" {
		t.Errorf("expected empty device, got %q", decoded.Device)
	}
}

func TestCompletionRequest_WithDevice(t *testing.T) {
	req := CompletionRequest{
		Model:  "phi-3-mini",
		Prompt: "hello",
		Device: "GPU",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var decoded CompletionRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Device != "GPU" {
		t.Errorf("expected GPU, got %q", decoded.Device)
	}
}

func TestCompletionRequest_WithoutDevice(t *testing.T) {
	req := CompletionRequest{
		Model:  "phi-3-mini",
		Prompt: "hello",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"model":"phi-3-mini","prompt":"hello"}` {
		t.Errorf("unexpected JSON with device omitted: %s", string(data))
	}
	var decoded CompletionRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Device != "" {
		t.Errorf("expected empty device, got %q", decoded.Device)
	}
}

func TestEmbeddingRequest_WithDevice(t *testing.T) {
	req := EmbeddingRequest{
		Model:  "all-MiniLM-L6-v2",
		Input:  []string{"hello world"},
		Device: "CPU",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var decoded EmbeddingRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Device != "CPU" {
		t.Errorf("expected CPU, got %q", decoded.Device)
	}
}

func TestEmbeddingRequest_WithoutDevice(t *testing.T) {
	req := EmbeddingRequest{
		Model: "all-MiniLM-L6-v2",
		Input: []string{"hello world"},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"model":"all-MiniLM-L6-v2","input":["hello world"]}` {
		t.Errorf("unexpected JSON with device omitted: %s", string(data))
	}
	var decoded EmbeddingRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Device != "" {
		t.Errorf("expected empty device, got %q", decoded.Device)
	}
}

func TestRerankRequest_WithDevice(t *testing.T) {
	req := RerankRequest{
		Model:     "ms-marco-MiniLM-L6-v2",
		Query:     "test query",
		Documents: []string{"doc1", "doc2"},
		Device:    "NPU",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var decoded RerankRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Device != "NPU" {
		t.Errorf("expected NPU, got %q", decoded.Device)
	}
}

func TestRerankRequest_WithoutDevice(t *testing.T) {
	req := RerankRequest{
		Model:     "ms-marco-MiniLM-L6-v2",
		Query:     "test query",
		Documents: []string{"doc1", "doc2"},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"model":"ms-marco-MiniLM-L6-v2","query":"test query","documents":["doc1","doc2"]}` {
		t.Errorf("unexpected JSON with device omitted: %s", string(data))
	}
	var decoded RerankRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Device != "" {
		t.Errorf("expected empty device, got %q", decoded.Device)
	}
}

func TestNewErrorResponse(t *testing.T) {
	t.Parallel()
	resp := NewErrorResponse("test_code", "test message")
	if resp.Error.Code != "test_code" {
		t.Errorf("expected test_code, got %q", resp.Error.Code)
	}
	if resp.Error.Message != "test message" {
		t.Errorf("expected 'test message', got %q", resp.Error.Message)
	}
}

func TestNewUsage(t *testing.T) {
	t.Parallel()
	u := newUsage(10, 20)
	if u.PromptTokens != 10 {
		t.Errorf("expected 10, got %d", u.PromptTokens)
	}
	if u.CompletionTokens != 20 {
		t.Errorf("expected 20, got %d", u.CompletionTokens)
	}
	if u.TotalTokens != 30 {
		t.Errorf("expected 30, got %d", u.TotalTokens)
	}
}

func TestUptime(t *testing.T) {
	t.Parallel()
	u := uptime()
	if u == "" {
		t.Error("expected non-empty uptime")
	}
}

func TestDeviceTypeConstants(t *testing.T) {
	t.Parallel()
	if DeviceCPU != DeviceType("cpu") {
		t.Errorf("expected cpu, got %q", DeviceCPU)
	}
	if DeviceGPU != DeviceType("gpu") {
		t.Errorf("expected gpu, got %q", DeviceGPU)
	}
	if DeviceNPU != DeviceType("npu") {
		t.Errorf("expected npu, got %q", DeviceNPU)
	}
}

func TestDeviceStruct(t *testing.T) {
	t.Parallel()
	d := Device{ID: "CPU", Name: "Intel CPU", Type: DeviceCPU, Available: true, Priority: 1}
	if d.ID != "CPU" {
		t.Error("ID mismatch")
	}
	if d.Type != DeviceCPU {
		t.Error("Type mismatch")
	}
	if !d.Available {
		t.Error("expected available")
	}
}

func TestModelInfoDefaults(t *testing.T) {
	t.Parallel()
	m := ModelInfo{ID: "test", Loaded: false}
	if m.ID != "test" {
		t.Error("ID mismatch")
	}
	if m.Loaded {
		t.Error("expected not loaded")
	}
}

func TestInferenceResult(t *testing.T) {
	t.Parallel()
	r := InferenceResult{Text: "hello", Tokens: []int{1, 2, 3}}
	if r.Text != "hello" {
		t.Errorf("expected hello, got %q", r.Text)
	}
	if len(r.Tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(r.Tokens))
	}
}

func TestTimingStruct(t *testing.T) {
	t.Parallel()
	tt := Timing{TTFT: "10ms", Total: "100ms", TokensPerSecond: 15.5, LoadDuration: "5s"}
	if tt.TTFT != "10ms" {
		t.Errorf("expected 10ms, got %q", tt.TTFT)
	}
	if tt.TokensPerSecond != 15.5 {
		t.Errorf("expected 15.5, got %f", tt.TokensPerSecond)
	}
}

func TestUsageStruct(t *testing.T) {
	t.Parallel()
	u := Usage{PromptTokens: 5, CompletionTokens: 10, TotalTokens: 15}
	if u.PromptTokens != 5 {
		t.Errorf("expected 5, got %d", u.PromptTokens)
	}
	if u.TotalTokens != 15 {
		t.Errorf("expected 15, got %d", u.TotalTokens)
	}
}

func TestBenchmarkResponseDefaults(t *testing.T) {
	t.Parallel()
	b := BenchmarkResponse{Model: "test", Device: "CPU"}
	if b.Model != "test" {
		t.Errorf("expected test, got %q", b.Model)
	}
	if b.TokensPerSecond != 0 {
		t.Errorf("expected 0, got %f", b.TokensPerSecond)
	}
}

func TestHealthResponse(t *testing.T) {
	t.Parallel()
	h := HealthResponse{Status: "ok", Version: "1.0", Uptime: "1s", ModelsLoaded: 2, ActiveDevice: "CPU"}
	if h.Status != "ok" {
		t.Errorf("expected ok, got %q", h.Status)
	}
	if h.ModelsLoaded != 2 {
		t.Errorf("expected 2, got %d", h.ModelsLoaded)
	}
}

func TestEmbeddingResult(t *testing.T) {
	t.Parallel()
	e := EmbeddingResult{Embeddings: [][]float32{{0.1, 0.2}, {0.3, 0.4}}}
	if len(e.Embeddings) != 2 {
		t.Errorf("expected 2 embeddings, got %d", len(e.Embeddings))
	}
	if e.Embeddings[0][0] != 0.1 {
		t.Errorf("expected 0.1, got %f", e.Embeddings[0][0])
	}
}

func TestChoiceDefaults(t *testing.T) {
	t.Parallel()
	c := Choice{Index: 0, Message: &Message{Role: "assistant", Content: "hi"}}
	if c.Message.Role != "assistant" {
		t.Errorf("expected assistant, got %q", c.Message.Role)
	}
	if c.FinishReason != nil {
		t.Error("expected nil finish reason")
	}
}
