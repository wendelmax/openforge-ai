//go:build cgo

package openvino

// Use pkg-config on Linux/macOS. On Windows, set CGO_CFLAGS/CGO_LDFLAGS env vars.

/*
#cgo !windows pkg-config: openvino

#include <openvino/c/openvino.h>
#include <stdlib.h>

// CGO cannot call C variadic functions directly.
static inline ov_status_e compile_model_no_props(ov_core_t* core, ov_model_t* model, const char* device, ov_compiled_model_t** compiled) {
    return ov_core_compile_model(core, model, device, 0, compiled);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func newCoreImpl() (*Core, error) {
	var core *C.ov_core_t
	status := C.ov_core_create(&core)
	if status != C.OK {
		return nil, fmt.Errorf("failed to create OpenVINO core: %s", Status(status).Error())
	}
	return &Core{ptr: unsafe.Pointer(core)}, nil
}

var NewCore = newCoreImpl

func NewTensor(elementType int32, shape []int64) (*Tensor, error) {
	var cShape C.ov_shape_t
	dims := make([]C.int64_t, len(shape))
	for i, d := range shape {
		dims[i] = C.int64_t(d)
	}
	rank := C.int64_t(len(shape))

	status := C.ov_shape_create(rank, &dims[0], &cShape)
	if status != C.OK {
		return nil, fmt.Errorf("failed to create shape: %s", Status(status).Error())
	}

	var tensor *C.ov_tensor_t
	status = C.ov_tensor_create(C.ov_element_type_e(elementType), cShape, &tensor)
	C.ov_shape_free(&cShape)
	if status != C.OK {
		return nil, fmt.Errorf("failed to create tensor: %s", Status(status).Error())
	}
	return &Tensor{ptr: unsafe.Pointer(tensor)}, nil
}

func (c *Core) ReadModel(path string) (*Model, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var model *C.ov_model_t
	status := C.ov_core_read_model((*C.ov_core_t)(c.ptr), cPath, nil, &model)
	if status != C.OK {
		return nil, fmt.Errorf("failed to read model %q: %s", path, Status(status).Error())
	}
	return &Model{ptr: unsafe.Pointer(model)}, nil
}

func (c *Core) CompileModel(model *Model, device string) (*CompiledModel, error) {
	cDevice := C.CString(device)
	defer C.free(unsafe.Pointer(cDevice))

	var compiled *C.ov_compiled_model_t
	status := C.compile_model_no_props((*C.ov_core_t)(c.ptr), (*C.ov_model_t)(model.ptr), cDevice, &compiled)
	if status != C.OK {
		return nil, fmt.Errorf("failed to compile model for device %q: %s", device, Status(status).Error())
	}
	return &CompiledModel{ptr: unsafe.Pointer(compiled)}, nil
}

func (c *Core) GetAvailableDevices() ([]string, error) {
	var devices C.ov_available_devices_t
	status := C.ov_core_get_available_devices((*C.ov_core_t)(c.ptr), &devices)
	if status != C.OK {
		return nil, fmt.Errorf("failed to get available devices: %s", Status(status).Error())
	}
	defer C.ov_available_devices_free(&devices)

	count := int(devices.size)
	result := make([]string, 0, count)
	devicePtrs := (*[1 << 30]*C.char)(unsafe.Pointer(devices.devices))[:count:count]
	for _, d := range devicePtrs {
		if d != nil {
			result = append(result, C.GoString(d))
		}
	}
	return result, nil
}

func (m *Model) InputsCount() (int64, error) {
	var count C.size_t
	status := C.ov_model_inputs_size((*C.ov_model_t)(m.ptr), &count)
	if status != C.OK {
		return 0, fmt.Errorf("failed to get inputs count: %s", Status(status).Error())
	}
	return int64(count), nil
}

func (m *Model) OutputsCount() (int64, error) {
	var count C.size_t
	status := C.ov_model_outputs_size((*C.ov_model_t)(m.ptr), &count)
	if status != C.OK {
		return 0, fmt.Errorf("failed to get outputs count: %s", Status(status).Error())
	}
	return int64(count), nil
}

func (m *Model) Free() {
	if m.ptr != nil {
		C.ov_model_free((*C.ov_model_t)(m.ptr))
		m.ptr = nil
	}
}

func (cm *CompiledModel) CreateInferRequest() (*InferRequest, error) {
	var req *C.ov_infer_request_t
	status := C.ov_compiled_model_create_infer_request((*C.ov_compiled_model_t)(cm.ptr), &req)
	if status != C.OK {
		return nil, fmt.Errorf("failed to create inference request: %s", Status(status).Error())
	}
	return &InferRequest{ptr: unsafe.Pointer(req)}, nil
}

func (cm *CompiledModel) Free() {
	if cm.ptr != nil {
		C.ov_compiled_model_free((*C.ov_compiled_model_t)(cm.ptr))
		cm.ptr = nil
	}
}

func (ir *InferRequest) SetInputTensor(idx int, tensor *Tensor) error {
	status := C.ov_infer_request_set_input_tensor_by_index(
		(*C.ov_infer_request_t)(ir.ptr), C.size_t(idx), (*C.ov_tensor_t)(tensor.ptr))
	if status != C.OK {
		return fmt.Errorf("failed to set input tensor %d: %s", idx, Status(status).Error())
	}
	return nil
}

func (ir *InferRequest) GetOutputTensor(idx int) (*Tensor, error) {
	var tensor *C.ov_tensor_t
	status := C.ov_infer_request_get_output_tensor_by_index(
		(*C.ov_infer_request_t)(ir.ptr), C.size_t(idx), &tensor)
	if status != C.OK {
		return nil, fmt.Errorf("failed to get output tensor %d: %s", idx, Status(status).Error())
	}
	return &Tensor{ptr: unsafe.Pointer(tensor)}, nil
}

func (ir *InferRequest) StartAsync() error {
	status := C.ov_infer_request_start_async((*C.ov_infer_request_t)(ir.ptr))
	if status != C.OK {
		return fmt.Errorf("failed to start async inference: %s", Status(status).Error())
	}
	return nil
}

func (ir *InferRequest) Wait() error {
	status := C.ov_infer_request_wait((*C.ov_infer_request_t)(ir.ptr))
	if status != C.OK {
		return fmt.Errorf("failed to wait for inference: %s", Status(status).Error())
	}
	return nil
}

func (ir *InferRequest) Infer() error {
	if err := ir.StartAsync(); err != nil {
		return err
	}
	return ir.Wait()
}

func (ir *InferRequest) Free() {
	if ir.ptr != nil {
		C.ov_infer_request_free((*C.ov_infer_request_t)(ir.ptr))
		ir.ptr = nil
	}
}

func (t *Tensor) Data() unsafe.Pointer {
	var data unsafe.Pointer
	C.ov_tensor_data((*C.ov_tensor_t)(t.ptr), &data)
	return data
}

func (t *Tensor) Free() {
	if t.ptr != nil {
		C.ov_tensor_free((*C.ov_tensor_t)(t.ptr))
		t.ptr = nil
	}
}

func (c *Core) Free() {
	if c.ptr != nil {
		C.ov_core_free((*C.ov_core_t)(c.ptr))
		c.ptr = nil
	}
}


