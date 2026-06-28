package openvino

import (
	"fmt"
	"unsafe"
)

type Status int32

const (
	StatusOK           Status = 0
	StatusGeneralError Status = -1
	StatusNotFound     Status = -2
	StatusInvalidParam Status = -3
	StatusBusy         Status = -4
	StatusUnsupported  Status = -5
)

const (
	ElementTypeI64 = 5
	ElementTypeF32 = 10
)

func (s Status) Error() string {
	switch s {
	case StatusOK:
		return "ok"
	case StatusGeneralError:
		return "general error"
	case StatusNotFound:
		return "not found"
	case StatusInvalidParam:
		return "invalid parameter"
	case StatusBusy:
		return "busy"
	case StatusUnsupported:
		return "unsupported"
	default:
		return fmt.Sprintf("unknown error (%d)", int(s))
	}
}

func (s Status) IsOK() bool {
	return s == StatusOK
}

type Core struct {
	ptr unsafe.Pointer
}

type Model struct {
	ptr unsafe.Pointer
}

type CompiledModel struct {
	ptr unsafe.Pointer
}

type InferRequest struct {
	ptr unsafe.Pointer
}

type Tensor struct {
	ptr unsafe.Pointer
}
