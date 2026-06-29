package openvino

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/runtime"
	"github.com/stretchr/testify/assert"
)

func TestNewRuntime(t *testing.T) {
	r := NewRuntime("/tmp/models")
	if r.models == nil {
		t.Fatal("expected non-nil models map")
	}
	if len(r.models) != 0 {
		t.Fatalf("expected empty models, got %d", len(r.models))
	}
}

func TestLoadModelStub(t *testing.T) {
	r := NewRuntime("/tmp/models")

	// Load a model in stub mode (core is nil)
	ctx := context.Background()
	err := r.LoadModel(ctx, "test-model", "/tmp/models/test-model", "CPU")
	if err != nil {
		t.Fatalf("LoadModel failed: %v", err)
	}

	lm, ok := r.models["test-model"]
	if !ok {
		t.Fatal("model not found in map after LoadModel")
	}
	if lm.info.ID != "test-model" {
		t.Fatalf("expected model ID 'test-model', got %q", lm.info.ID)
	}
	if !lm.info.Loaded {
		t.Fatal("expected model to be marked as loaded")
	}
}

func TestLoadModelIdempotentStub(t *testing.T) {
	r := NewRuntime("/tmp/models")
	ctx := context.Background()

	err := r.LoadModel(ctx, "test-model", "/tmp/models/test-model", "CPU")
	if err != nil {
		t.Fatalf("first LoadModel failed: %v", err)
	}

	// Loading the same model again should succeed (idempotent)
	err = r.LoadModel(ctx, "test-model", "/tmp/models/test-model", "CPU")
	if err != nil {
		t.Fatalf("second LoadModel should be idempotent, got: %v", err)
	}
}

func TestLoadModelMultipleDevicesStub(t *testing.T) {
	r := NewRuntime("/tmp/models")
	ctx := context.Background()

	err := r.LoadModel(ctx, "test-model", "/tmp/models/test-model", "CPU")
	if err != nil {
		t.Fatalf("first LoadModel failed: %v", err)
	}

	// Loading same model for different device in stub mode creates a new stub entry
	err = r.LoadModel(ctx, "test-model", "/tmp/models/test-model", "GPU")
	if err != nil {
		t.Fatalf("LoadModel for second device failed: %v", err)
	}

	lm, ok := r.models["test-model"]
	if !ok {
		t.Fatal("model not found")
	}
	if !lm.info.Loaded {
		t.Fatal("model should be loaded")
	}
}

func TestUnloadModelStub(t *testing.T) {
	r := NewRuntime("/tmp/models")
	ctx := context.Background()

	err := r.LoadModel(ctx, "test-model", "/tmp/models/test-model", "CPU")
	if err != nil {
		t.Fatalf("LoadModel failed: %v", err)
	}

	err = r.UnloadModel(ctx, "test-model")
	if err != nil {
		t.Fatalf("UnloadModel failed: %v", err)
	}

	if _, ok := r.models["test-model"]; ok {
		t.Fatal("model should be removed after UnloadModel")
	}
}

func TestUnloadModelNotLoaded(t *testing.T) {
	r := NewRuntime("/tmp/models")
	ctx := context.Background()

	err := r.UnloadModel(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error when unloading nonexistent model")
	}
}

func TestCloseStub(t *testing.T) {
	r := NewRuntime("/tmp/models")
	ctx := context.Background()

	_ = r.LoadModel(ctx, "model-a", "/tmp/models/model-a", "CPU")
	_ = r.LoadModel(ctx, "model-b", "/tmp/models/model-b", "CPU")

	err := r.Close(ctx)
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if len(r.models) != 0 {
		t.Fatalf("expected models to be cleared after Close, got %d", len(r.models))
	}
}

func TestCompiledModelHandleType(t *testing.T) {
	// Verify the compiledModelHandle struct exists with expected fields
	h := &compiledModelHandle{
		device: "CPU",
	}
	if h.device != "CPU" {
		t.Fatalf("expected device 'CPU', got %q", h.device)
	}
	if h.compiled != nil {
		t.Fatal("expected nil compiled for uninitialized handle")
	}
}

func TestLoadedModelStruct(t *testing.T) {
	// Verify loadedModel struct works with compiledByDevice map
	lm := &loadedModel{
		compiledByDevice: make(map[string]*compiledModelHandle),
	}
	lm.compiledByDevice["CPU"] = &compiledModelHandle{device: "CPU"}
	lm.compiledByDevice["GPU"] = &compiledModelHandle{device: "GPU"}

	if len(lm.compiledByDevice) != 2 {
		t.Fatalf("expected 2 compiled devices, got %d", len(lm.compiledByDevice))
	}
	if _, ok := lm.compiledByDevice["CPU"]; !ok {
		t.Fatal("expected CPU device in map")
	}
	if _, ok := lm.compiledByDevice["GPU"]; !ok {
		t.Fatal("expected GPU device in map")
	}
}

func TestDeviceForWorkload_RequestDevice(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.defaultDev = "CPU"

	result := r.DeviceForWorkload("chat", "NPU")
	if result != "NPU" {
		t.Errorf("expected NPU, got %q", result)
	}
}

func TestDeviceForWorkload_WorkloadDefault(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.defaultDev = "CPU"
	r.deviceConfig = &config.DeviceConfig{
		Chat: "GPU",
	}

	result := r.DeviceForWorkload("chat", "")
	if result != "GPU" {
		t.Errorf("expected GPU, got %q", result)
	}
}

func TestDeviceForWorkload_GlobalDefault(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.defaultDev = "NPU"

	result := r.DeviceForWorkload("chat", "")
	if result != "NPU" {
		t.Errorf("expected NPU, got %q", result)
	}
}

func TestDeviceForWorkload_FallbackCPU(t *testing.T) {
	r := NewRuntime("/tmp/models")

	result := r.DeviceForWorkload("chat", "")
	if result != "CPU" {
		t.Errorf("expected CPU, got %q", result)
	}
}

func TestNewProvider(t *testing.T) {
	p := NewProvider("/tmp/models")
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.Name() != "openvino" {
		t.Errorf("expected name openvino, got %q", p.Name())
	}
	if p.Runtime() == nil {
		t.Fatal("expected non-nil runtime")
	}
}

func TestProviderInitializeStub(t *testing.T) {
	p := NewProvider("/tmp/models")
	err := p.Initialize(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestProviderShutdown(t *testing.T) {
	p := NewProvider("/tmp/models")
	_ = p.Initialize(context.Background())
	err := p.Shutdown(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetDefaultDevice(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.SetDefaultDevice("NPU")
	r.mu.RLock()
	if r.defaultDev != "NPU" {
		t.Errorf("expected NPU, got %q", r.defaultDev)
	}
	r.mu.RUnlock()
}

func TestSetDeviceConfig(t *testing.T) {
	r := NewRuntime("/tmp/models")
	cfg := &config.DeviceConfig{Chat: "GPU", Default: "CPU", Embedding: "NPU", Rerank: "CPU"}
	r.SetDeviceConfig(cfg)
	r.mu.RLock()
	if r.deviceConfig.Chat != "GPU" {
		t.Errorf("expected GPU, got %q", r.deviceConfig.Chat)
	}
	r.mu.RUnlock()
}

func TestSetTokenizer(t *testing.T) {
	r := NewRuntime("/tmp/models")
	tok := NewByteLevelTokenizer(1000)
	r.SetTokenizer(tok)
	r.mu.RLock()
	if r.tokenizer == nil {
		t.Fatal("tokenizer should not be nil")
	}
	r.mu.RUnlock()
}

func TestInitializeSetsDefaultTokenizer(t *testing.T) {
	r := NewRuntime("/tmp/models")
	_ = r.Initialize(context.Background())
	r.mu.RLock()
	if r.tokenizer == nil {
		t.Fatal("expected tokenizer after init")
	}
	r.mu.RUnlock()
}

func TestInitializeIdempotent(t *testing.T) {
	r := NewRuntime("/tmp/models")
	err1 := r.Initialize(context.Background())
	err2 := r.Initialize(context.Background())
	if err1 != nil || err2 != nil {
		t.Fatal("Initialize should be idempotent")
	}
}

func TestDeviceForWorkload_Auto(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.defaultDev = "NPU"
	result := r.DeviceForWorkload("chat", "auto")
	if result != "NPU" {
		t.Errorf("expected NPU (global default), got %q", result)
	}
}

func TestDeviceForWorkload_EmptyRequestDevice(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.defaultDev = "GPU"
	result := r.DeviceForWorkload("chat", "")
	if result != "GPU" {
		t.Errorf("expected GPU, got %q", result)
	}
}

func TestDeviceForWorkload_WorkloadCompletion(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.deviceConfig = &config.DeviceConfig{Default: "GPU"}
	result := r.DeviceForWorkload("completion", "")
	if result != "GPU" {
		t.Errorf("expected GPU, got %q", result)
	}
}

func TestDeviceForWorkload_WorkloadEmbedding(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.deviceConfig = &config.DeviceConfig{Embedding: "NPU"}
	result := r.DeviceForWorkload("embedding", "")
	if result != "NPU" {
		t.Errorf("expected NPU, got %q", result)
	}
}

func TestDeviceForWorkload_WorkloadRerank(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.deviceConfig = &config.DeviceConfig{Rerank: "GPU"}
	result := r.DeviceForWorkload("rerank", "")
	if result != "GPU" {
		t.Errorf("expected GPU, got %q", result)
	}
}

func TestDeviceForWorkload_WorkloadChat(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.deviceConfig = &config.DeviceConfig{Chat: "NPU"}
	result := r.DeviceForWorkload("chat", "")
	if result != "NPU" {
		t.Errorf("expected NPU, got %q", result)
	}
}

func TestListDevicesStub(t *testing.T) {
	r := NewRuntime("/tmp/models")
	devices, err := r.ListDevices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 stub device, got %d", len(devices))
	}
	if devices[0].ID != "stub" {
		t.Errorf("expected 'stub', got %q", devices[0].ID)
	}
}

func TestListModelsEmptyDir(t *testing.T) {
	r := NewRuntime("/tmp/models")
	models, err := r.ListModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 0 {
		t.Errorf("expected 0 models, got %d", len(models))
	}
}

func TestCategorizeDevice(t *testing.T) {
	tests := []struct {
		name string
		want runtime.DeviceType
	}{
		{"CPU", runtime.DeviceCPU},
		{"GPU", runtime.DeviceGPU},
		{"NPU", runtime.DeviceNPU},
		{"CPU_A", runtime.DeviceCPU},
		{"GPU.0", runtime.DeviceGPU},
		{"NPU_1", runtime.DeviceNPU},
		{"UNKNOWN", runtime.DeviceCPU},
		{"", runtime.DeviceCPU},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := categorizeDevice(tt.name); got != tt.want {
				t.Errorf("categorizeDevice(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestSelectBestDevice(t *testing.T) {
	devices := []runtime.Device{
		{ID: "CPU", Type: runtime.DeviceCPU, Available: true, Priority: 1},
		{ID: "GPU", Type: runtime.DeviceGPU, Available: true, Priority: 3},
		{ID: "NPU", Type: runtime.DeviceNPU, Available: true, Priority: 2},
	}
	best := selectBestDevice(devices)
	if best != "GPU" {
		t.Errorf("expected GPU (highest priority), got %q", best)
	}
}

func TestSelectBestDevice_OnlyCPU(t *testing.T) {
	devices := []runtime.Device{
		{ID: "CPU", Type: runtime.DeviceCPU},
	}
	best := selectBestDevice(devices)
	if best != "CPU" {
		t.Errorf("expected CPU, got %q", best)
	}
}

func TestSelectBestDevice_Empty(t *testing.T) {
	best := selectBestDevice(nil)
	if best != "CPU" {
		t.Errorf("expected CPU fallback, got %q", best)
	}
}

func TestSelectBestDevice_UnknownType(t *testing.T) {
	devices := []runtime.Device{
		{ID: "CUSTOM", Type: runtime.DeviceType("custom")},
	}
	best := selectBestDevice(devices)
	if best != "CPU" {
		t.Errorf("expected CPU for unknown type, got %q", best)
	}
}

func TestSelectBestDevice_OnlyNPU(t *testing.T) {
	devices := []runtime.Device{
		{ID: "NPU", Type: runtime.DeviceNPU},
	}
	best := selectBestDevice(devices)
	if best != "NPU" {
		t.Errorf("expected NPU, got %q", best)
	}
}

func TestStatusError(t *testing.T) {
	tests := []struct {
		s    Status
		want string
	}{
		{StatusOK, "ok"},
		{StatusGeneralError, "general error"},
		{StatusNotFound, "not found"},
		{StatusInvalidParam, "invalid parameter"},
		{StatusBusy, "busy"},
		{StatusUnsupported, "unsupported"},
		{Status(99), "unknown error (99)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.s.Error(); got != tt.want {
				t.Errorf("Status(%d).Error() = %q, want %q", int(tt.s), got, tt.want)
			}
		})
	}
}

func TestStatusIsOK(t *testing.T) {
	if !StatusOK.IsOK() {
		t.Error("expected StatusOK.IsOK() to be true")
	}
	if StatusGeneralError.IsOK() {
		t.Error("expected StatusGeneralError.IsOK() to be false")
	}
}

func TestByteLevelTokenizer_EncodeEdgeCases(t *testing.T) {
	tok := NewByteLevelTokenizer(256)
	ids, err := tok.Encode("abc")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 {
		t.Errorf("expected 3 ids, got %d", len(ids))
	}
}

func TestWhitespaceTokenizer_Empty(t *testing.T) {
	tok := NewWhitespaceTokenizer(50000)
	ids, err := tok.Encode("")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Errorf("expected empty, got %d", len(ids))
	}
}

func TestWhitespaceTokenizer_UnknownToken(t *testing.T) {
	tok := NewWhitespaceTokenizer(50000)
	ids, err := tok.Encode("hello hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 ids, got %d", len(ids))
	}
	if ids[0] != ids[1] {
		t.Errorf("same word should produce same id")
	}
}

func TestCompositeTokenizer(t *testing.T) {
	tok := &CompositeTokenizer{
		tokenizers: []Tokenizer{
			NewByteLevelTokenizer(50272),
			NewWhitespaceTokenizer(50000),
		},
	}
	ids, err := tok.Encode("hello world")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) == 0 {
		t.Error("expected non-empty ids")
	}
	text, err := tok.Decode(ids)
	if err != nil {
		t.Fatal(err)
	}
	if text == "" {
		t.Error("expected non-empty decoded text")
	}
}

func TestCompositeTokenizer_VocabSize(t *testing.T) {
	tok := &CompositeTokenizer{}
	if tok.VocabSize() != 0 {
		t.Errorf("expected 0 for empty composite, got %d", tok.VocabSize())
	}
	tok2 := &CompositeTokenizer{
		tokenizers: []Tokenizer{NewByteLevelTokenizer(50272)},
	}
	if tok2.VocabSize() != 50272 {
		t.Errorf("expected 50272, got %d", tok2.VocabSize())
	}
}

func TestCompositeTokenizer_EncodeError(t *testing.T) {
	errorTok := &errTokenizer{}
	tok := &CompositeTokenizer{
		tokenizers: []Tokenizer{errorTok},
	}
	_, err := tok.Encode("test")
	if err == nil {
		t.Error("expected error from errTokenizer")
	}
}

func TestCompositeTokenizer_DecodeError(t *testing.T) {
	errorTok := &errTokenizer{}
	tok := &CompositeTokenizer{
		tokenizers: []Tokenizer{errorTok},
	}
	_, err := tok.Decode([]int64{1, 2, 3})
	if err == nil {
		t.Error("expected error from errTokenizer")
	}
}

type errTokenizer struct{}

func (e *errTokenizer) Encode(text string) ([]int64, error) {
	return nil, assert.AnError
}
func (e *errTokenizer) Decode(tokens []int64) (string, error) {
	return "", assert.AnError
}
func (e *errTokenizer) VocabSize() int { return 0 }

func TestStubInferReturnsError(t *testing.T) {
	r := NewRuntime("/tmp/models")
	_ = r.Initialize(context.Background())
	_, err := r.Infer(context.Background(), "nonexistent", "test", runtime.InferenceParams{})
	if err == nil {
		t.Error("expected error for nonexistent model in stub mode")
	}
}

func TestStubInferStreamReturnsError(t *testing.T) {
	r := NewRuntime("/tmp/models")
	_ = r.Initialize(context.Background())
	_, err := r.InferStream(context.Background(), "nonexistent", "test", runtime.InferenceParams{})
	if err == nil {
		t.Error("expected error for nonexistent model in stub mode")
	}
}

func TestStubEmbedReturnsError(t *testing.T) {
	r := NewRuntime("/tmp/models")
	_ = r.Initialize(context.Background())
	_, err := r.Embed(context.Background(), "nonexistent", []string{"test"}, "")
	if err == nil {
		t.Error("expected error for nonexistent model in stub mode")
	}
}

func TestStubBenchmarkReturnsError(t *testing.T) {
	r := NewRuntime("/tmp/models")
	_ = r.Initialize(context.Background())
	_, err := r.Benchmark(context.Background(), "nonexistent", 1, "test", 10)
	if err == nil {
		t.Error("expected error for nonexistent model in stub mode")
	}
}

func TestInitializeNonStubDiscoverDevicesSuccess(t *testing.T) {
	if _, err := NewCore(); err == nil {
		t.Skip("real OpenVINO available; stub mock test skipped")
	}
	origNewCore := NewCore
	NewCore = func() (*Core, error) { return &Core{}, nil }
	defer func() { NewCore = origNewCore }()

	origGetDevices := stubGetAvailableDevices
	stubGetAvailableDevices = func() ([]string, error) {
		return []string{"CPU", "GPU"}, nil
	}
	defer func() { stubGetAvailableDevices = origGetDevices }()

	r := NewRuntime("/tmp/models")
	err := r.Initialize(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	r.mu.RLock()
	if r.defaultDev != "GPU" {
		t.Errorf("expected GPU (highest priority), got %q", r.defaultDev)
	}
	r.mu.RUnlock()
}

func TestInitializeNonStubPath(t *testing.T) {
	origNewCore := NewCore
	NewCore = func() (*Core, error) { return &Core{}, nil }
	defer func() { NewCore = origNewCore }()

	r := NewRuntime("/tmp/models")
	err := r.Initialize(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	r.mu.RLock()
	if r.defaultDev == "stub" {
		t.Error("expected non-stub default device")
	}
	if r.tokenizer == nil {
		t.Error("expected tokenizer to be set")
	}
	r.mu.RUnlock()
}

func TestInitializeSetsStubDevice(t *testing.T) {
	if _, err := NewCore(); err == nil {
		t.Skip("real OpenVINO available; stub test skipped")
	}
	r := NewRuntime("/tmp/models")
	_ = r.Initialize(context.Background())
	r.mu.RLock()
	if r.defaultDev != "stub" {
		t.Errorf("expected 'stub' default device, got %q", r.defaultDev)
	}
	r.mu.RUnlock()
}

func TestFindModelFile_NonExistent(t *testing.T) {
	path := findModelFile("/nonexistent/path")
	if path != "" {
		t.Errorf("expected empty string, got %q", path)
	}
}

func TestConstants(t *testing.T) {
	if ElementTypeI64 != 5 {
		t.Errorf("expected 5, got %d", ElementTypeI64)
	}
	if ElementTypeF32 != 10 {
		t.Errorf("expected 10, got %d", ElementTypeF32)
	}
}

func TestStubNewTensor(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestStubReadModel(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestStubCompileModel(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestStubGetAvailableDevices(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestStubFree(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestStubInputsCount(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestStubOutputsCount(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestStubCreateInferRequest(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestStubSetInputTensor(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestStubGetOutputTensor(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestStubInferRequestMethods(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestStubTensorData(t *testing.T) {
	t.Skip("requires stub build tag; CGO runtime is active")
}

func TestLoadModel_AutoDevice(t *testing.T) {
	t.Skip("requires real model path on filesystem")
}

func TestWhitespaceTokenizer_VocabSize(t *testing.T) {
	t.Parallel()
	tok := NewWhitespaceTokenizer(50000)
	if tok.VocabSize() != 50000 {
		t.Errorf("expected 50000, got %d", tok.VocabSize())
	}
}

func TestWhitespaceTokenizer_DecodeUnknown(t *testing.T) {
	t.Parallel()
	tok := NewWhitespaceTokenizer(50000)
	text, err := tok.Decode([]int64{99999})
	if err != nil {
		t.Fatal(err)
	}
	if text != "" {
		t.Errorf("expected empty string, got %q", text)
	}
}

func TestFindModelFile_XmlPath(t *testing.T) {
	dir := t.TempDir()
	xmlFile := filepath.Join(dir, "model.xml")
	if err := os.WriteFile(xmlFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	path := findModelFile(xmlFile)
	if path != xmlFile {
		t.Errorf("expected %q, got %q", xmlFile, path)
	}
}

func TestFindModelFile_GlobMatch(t *testing.T) {
	dir := t.TempDir()
	modelDir := filepath.Join(dir, "model")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatal(err)
	}
	xmlFile := filepath.Join(modelDir, "openvino_model.xml")
	if err := os.WriteFile(xmlFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	path := findModelFile(modelDir)
	if path != xmlFile {
		t.Errorf("expected %q, got %q", xmlFile, path)
	}
}

func TestFindModelFile_DirNoXML(t *testing.T) {
	dir := t.TempDir()
	path := findModelFile(dir)
	if path != "" {
		t.Errorf("expected empty for dir without XML, got %q", path)
	}
}

func TestFindModelFile_NonXMLFile(t *testing.T) {
	dir := t.TempDir()
	txtFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(txtFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	path := findModelFile(txtFile)
	if path != "" {
		t.Errorf("expected empty for non-XML file, got %q", path)
	}
}

func TestFindModelFile_MultiGlobNoOpenVINO(t *testing.T) {
	dir := t.TempDir()
	modelDir := filepath.Join(dir, "model")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatal(err)
	}
	xmlFile1 := filepath.Join(modelDir, "model.xml")
	xmlFile2 := filepath.Join(modelDir, "other.xml")
	os.WriteFile(xmlFile1, []byte("test"), 0644)
	os.WriteFile(xmlFile2, []byte("test"), 0644)
	path := findModelFile(modelDir)
	if path != xmlFile1 {
		t.Errorf("expected %q, got %q", xmlFile1, path)
	}
}

func TestFindModelFile_MultiGlobPrefersOpenVINO(t *testing.T) {
	dir := t.TempDir()
	modelDir := filepath.Join(dir, "model")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatal(err)
	}
	xmlFile1 := filepath.Join(modelDir, "model.xml")
	xmlFile2 := filepath.Join(modelDir, "openvino_model.xml")
	os.WriteFile(xmlFile1, []byte("test"), 0644)
	os.WriteFile(xmlFile2, []byte("test"), 0644)
	path := findModelFile(modelDir)
	if path != xmlFile2 {
		t.Errorf("expected openvino_model.xml, got %q", path)
	}
}

func TestListModels_WithDir(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "model-a"), 0755)
	os.MkdirAll(filepath.Join(dir, "model-b"), 0755)
	r := NewRuntime(dir)
	models, err := r.ListModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 {
		t.Errorf("expected 2 models, got %d", len(models))
	}
}

func TestListModels_WithLoadedModel(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "test-model"), 0755)
	r := NewRuntime(dir)
	_ = r.LoadModel(context.Background(), "test-model", filepath.Join(dir, "test-model"), "CPU")
	models, _ := r.ListModels(context.Background())
	if len(models) < 1 {
		t.Fatal("expected at least 1 model")
	}
	for _, m := range models {
		if m.ID == "test-model" && !m.Loaded {
			t.Error("expected test-model to be marked as loaded")
		}
	}
}

func TestCloseReleasesModels(t *testing.T) {
	r := NewRuntime("/tmp/models")
	_ = r.LoadModel(context.Background(), "test-model", "test-path", "CPU")
	_ = r.LoadModel(context.Background(), "model-b", "path-b", "GPU")
	if len(r.models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(r.models))
	}
	_ = r.Close(context.Background())
	if len(r.models) != 0 {
		t.Errorf("expected 0 models after Close, got %d", len(r.models))
	}
}

func TestUnloadModelWithModelFree(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.mu.Lock()
	r.models["test"] = &loadedModel{
		info: runtime.ModelInfo{ID: "test", Loaded: true},
		model: &Model{},
		compiledByDevice: map[string]*compiledModelHandle{
			"CPU": {compiled: &CompiledModel{}, device: "CPU"},
		},
	}
	r.mu.Unlock()
	err := r.UnloadModel(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.models["test"]; ok {
		t.Error("model should be removed after UnloadModel")
	}
}

func TestCloseWithCoreFree(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.core = &Core{}
	r.models["m"] = &loadedModel{info: runtime.ModelInfo{ID: "m", Loaded: true}}
	err := r.Close(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if r.core != nil {
		t.Error("expected core to be nil after Close")
	}
}

func TestListDevicesWithCore(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.core = &Core{}
	devices, err := r.ListDevices(context.Background())
	if err == nil {
		t.Error("expected error in stub mode with core set")
	}
	if devices != nil {
		t.Error("expected nil devices on error")
	}
}

func TestLoadModelNonStubFileNotFound(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.core = &Core{}
	err := r.LoadModel(context.Background(), "test", "/nonexistent", "CPU")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestLoadModelNonStubReadModelError(t *testing.T) {
	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "model.xml")
	if err := os.WriteFile(xmlPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	r := NewRuntime("/tmp/models")
	r.core = &Core{}
	err := r.LoadModel(context.Background(), "test", dir, "CPU")
	if err == nil {
		t.Error("expected error from stub ReadModel")
	}
}

func TestLoadModelNonStubAutoDevice(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.core = &Core{}
	err := r.LoadModel(context.Background(), "test", "/nonexistent", "auto")
	if err == nil {
		t.Error("expected error for auto device in non-stub mode")
	}
}

func TestLoadModelNonStubExistingModelNewDevice(t *testing.T) {
	if _, err := NewCore(); err == nil {
		t.Skip("real OpenVINO CGO available; CompileModel nil in test path")
	}
	r := NewRuntime("/tmp/models")
	_ = r.LoadModel(context.Background(), "test", "/tmp/path", "CPU")
	r.core = &Core{}
	err := r.LoadModel(context.Background(), "test", "/tmp/path", "GPU")
	if err == nil {
		t.Error("expected error for compilation in stub mode")
	}
}

func TestLoadModelNonStubRelativePath(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.core = &Core{}
	err := r.LoadModel(context.Background(), "test", "relative/path", "CPU")
	if err == nil {
		t.Error("expected error for non-existent relative path")
	}
}

func TestLoadModelNonStubIdempotent(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.mu.Lock()
	r.models["test"] = &loadedModel{
		info: runtime.ModelInfo{ID: "test", Loaded: true},
		model: &Model{},
		compiledByDevice: map[string]*compiledModelHandle{
			"CPU": {compiled: &CompiledModel{}, device: "CPU"},
		},
	}
	r.mu.Unlock()
	r.core = &Core{}
	err := r.LoadModel(context.Background(), "test", "/tmp/path", "CPU")
	if err != nil {
		t.Fatal(err)
	}
}

func TestListModelsErrorWithCore(t *testing.T) {
	r := NewRuntime("/nonexistent-test-path")
	r.core = &Core{}
	_, err := r.ListModels(context.Background())
	if err == nil {
		t.Error("expected error for nonexistent path with core set")
	}
}

func TestCloseWithCompiledDevices(t *testing.T) {
	r := NewRuntime("/tmp/models")
	r.mu.Lock()
	r.models["test"] = &loadedModel{
		info: runtime.ModelInfo{ID: "test", Loaded: true},
		model: &Model{},
		compiledByDevice: map[string]*compiledModelHandle{
			"CPU": {compiled: &CompiledModel{}, device: "CPU"},
		},
	}
	r.mu.Unlock()
	_ = r.Close(context.Background())
	if len(r.models) != 0 {
		t.Errorf("expected 0 models after Close, got %d", len(r.models))
	}
}

func TestLoadModel_AutoDevice_NonStub(t *testing.T) {
	t.Skip("requires real model path on filesystem")
}

func TestLoadModelAutoDeviceSkip(t *testing.T) {
	t.Skip("duplicate of skipped test")
}
