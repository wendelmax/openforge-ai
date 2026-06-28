//go:build !cgo

package openvino

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/openforge-ai/openforge/runtime"
)

var NewCore = func() (*Core, error) {
	return nil, fmt.Errorf("OpenVINO requires CGO: rebuild with CGO_ENABLED=1 and OpenVINO installed")
}

func NewTensor(elementType int32, shape []int64) (*Tensor, error) {
	return nil, fmt.Errorf("OpenVINO requires CGO: rebuild with CGO_ENABLED=1 and OpenVINO installed")
}

func (c *Core) ReadModel(path string) (*Model, error) {
	return nil, fmt.Errorf("OpenVINO requires CGO")
}

func (c *Core) CompileModel(model *Model, device string) (*CompiledModel, error) {
	return nil, fmt.Errorf("OpenVINO requires CGO")
}

var stubGetAvailableDevices func() ([]string, error)

func (c *Core) GetAvailableDevices() ([]string, error) {
	if stubGetAvailableDevices != nil {
		return stubGetAvailableDevices()
	}
	return nil, fmt.Errorf("OpenVINO requires CGO")
}

func (c *Core) Free() {}

func (m *Model) InputsCount() (int64, error) { return 0, nil }

func (m *Model) OutputsCount() (int64, error) { return 0, nil }

func (m *Model) Free() {}

func (cm *CompiledModel) CreateInferRequest() (*InferRequest, error) {
	return nil, fmt.Errorf("OpenVINO requires CGO")
}

func (cm *CompiledModel) Free() {}

func (ir *InferRequest) SetInputTensor(idx int, tensor *Tensor) error {
	return fmt.Errorf("OpenVINO requires CGO")
}

func (ir *InferRequest) GetOutputTensor(idx int) (*Tensor, error) {
	return nil, fmt.Errorf("OpenVINO requires CGO")
}

func (ir *InferRequest) StartAsync() error { return fmt.Errorf("OpenVINO requires CGO") }

func (ir *InferRequest) Wait() error { return fmt.Errorf("OpenVINO requires CGO") }

func (ir *InferRequest) Infer() error { return fmt.Errorf("OpenVINO requires CGO") }

func (ir *InferRequest) Free() {}

func (t *Tensor) Data() unsafe.Pointer { return nil }

func (t *Tensor) Free() {}

var ovElementTypeI64 int32 = 5
var ovElementTypeF32 int32 = 10

func (r *OpenVINORuntime) Infer(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (*runtime.InferenceResult, error) {
	return nil, fmt.Errorf("OpenVINO requires CGO: rebuild with CGO_ENABLED=1")
}

func (r *OpenVINORuntime) InferStream(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (<-chan string, error) {
	return nil, fmt.Errorf("OpenVINO requires CGO: rebuild with CGO_ENABLED=1")
}

func (r *OpenVINORuntime) Embed(ctx context.Context, modelID string, inputs []string, device string) (*runtime.EmbeddingResult, error) {
	return nil, fmt.Errorf("OpenVINO requires CGO: rebuild with CGO_ENABLED=1")
}

func (r *OpenVINORuntime) Benchmark(ctx context.Context, modelID string, iterations int, prompt string, maxTokens int) (runtime.BenchmarkResults, error) {
	return nil, fmt.Errorf("OpenVINO requires CGO: rebuild with CGO_ENABLED=1")
}

func init() {
	_ = unsafe.Pointer(nil)
}
