package openvino

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/openforge-ai/openforge/internal/cache"
	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/runtime"
)

type OpenVINORuntime struct {
	mu          sync.RWMutex
	core        *Core
	models      map[string]*loadedModel
	tokenizer   Tokenizer
	modelPath   string
	defaultDev  string
	deviceConfig *config.DeviceConfig
	initialized  bool
	embedCache   *cache.Cache
}

type compiledModelHandle struct {
	compiled *CompiledModel
	device   string
}

type loadedModel struct {
	info             runtime.ModelInfo
	model            *Model
	compiledByDevice map[string]*compiledModelHandle
}

func NewRuntime(modelPath string) *OpenVINORuntime {
	return &OpenVINORuntime{
		models:     make(map[string]*loadedModel),
		modelPath:  modelPath,
		embedCache: cache.New(),
	}
}

func (r *OpenVINORuntime) SetDefaultDevice(device string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultDev = device
}

func (r *OpenVINORuntime) SetDeviceConfig(cfg *config.DeviceConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deviceConfig = cfg
}

func (r *OpenVINORuntime) DeviceForWorkload(workload, requestDevice string) string {
	if requestDevice != "" && requestDevice != "auto" {
		return requestDevice
	}
	r.mu.RLock()
	devCfg := r.deviceConfig
	defDev := r.defaultDev
	r.mu.RUnlock()
	if devCfg != nil {
		var wlDevice string
		switch workload {
		case "chat":
			wlDevice = devCfg.Chat
		case "completion":
			wlDevice = devCfg.Default
		case "embedding":
			wlDevice = devCfg.Embedding
		case "rerank":
			wlDevice = devCfg.Rerank
		}
		if wlDevice != "" {
			return wlDevice
		}
	}
	if defDev != "" {
		return defDev
	}
	return "CPU"
}

func (r *OpenVINORuntime) SetTokenizer(t Tokenizer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokenizer = t
}

func (r *OpenVINORuntime) Initialize(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.initialized {
		return nil
	}

	core, err := NewCore()
	if err != nil {
		slog.Warn("OpenVINO core not available, running in stub mode", "error", err)
		r.initialized = true
		r.defaultDev = "stub"
		if r.tokenizer == nil {
			r.tokenizer = NewByteLevelTokenizer(50272)
		}
		return nil
	}

	r.core = core
	r.initialized = true

	devices, err := r.discoverDevices(ctx)
	if err != nil {
		slog.Warn("device discovery incomplete", "error", err)
	} else {
		slog.Info("devices detected", "count", len(devices))
		for _, d := range devices {
			slog.Debug("device available", "id", d.ID, "type", d.Type)
		}
	}

	if r.defaultDev == "" || r.defaultDev == "auto" {
		r.defaultDev = selectBestDevice(devices)
		slog.Info("auto-selected device", "device", r.defaultDev)
	}

	if r.tokenizer == nil {
		r.tokenizer = NewByteLevelTokenizer(50272)
		slog.Warn("using default byte-level tokenizer; quality may be limited")
	}

	return nil
}

func (r *OpenVINORuntime) discoverDevices(ctx context.Context) ([]runtime.Device, error) {
	deviceNames, err := r.core.GetAvailableDevices()
	if err != nil {
		return nil, err
	}

	devices := make([]runtime.Device, 0, len(deviceNames))
	for _, name := range deviceNames {
		upper := strings.ToUpper(name)
		dev := runtime.Device{
			ID:        name,
			Name:      name,
			Type:      categorizeDevice(upper),
			Available: true,
		}
		devices = append(devices, dev)
	}

	return devices, nil
}

func categorizeDevice(name string) runtime.DeviceType {
	switch {
	case strings.HasPrefix(name, "CPU"):
		return runtime.DeviceCPU
	case strings.HasPrefix(name, "GPU"):
		return runtime.DeviceGPU
	case strings.HasPrefix(name, "NPU"):
		return runtime.DeviceNPU
	default:
		return runtime.DeviceCPU
	}
}

func selectBestDevice(devices []runtime.Device) string {
	priority := func(t runtime.DeviceType) int {
		switch t {
		case runtime.DeviceGPU:
			return 3
		case runtime.DeviceNPU:
			return 2
		case runtime.DeviceCPU:
			return 1
		default:
			return 0
		}
	}

	best := "CPU"
	bestPrio := 0

	for _, d := range devices {
		if p := priority(d.Type); p > bestPrio {
			bestPrio = p
			best = d.ID
		}
	}
	return best
}

func (r *OpenVINORuntime) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, lm := range r.models {
		for _, ch := range lm.compiledByDevice {
			ch.compiled.Free()
		}
		if lm.model != nil {
			lm.model.Free()
		}
		delete(r.models, id)
		slog.Debug("model unloaded", "model_id", id)
	}

	if r.core != nil {
		r.core.Free()
		r.core = nil
	}

	r.initialized = false
	return nil
}

func (r *OpenVINORuntime) ListDevices(ctx context.Context) ([]runtime.Device, error) {
	if r.core == nil {
		return []runtime.Device{
			{ID: "stub", Name: "Stub Mode (no OpenVINO)", Type: runtime.DeviceCPU, Available: true, Priority: 1},
		}, nil
	}
	return r.discoverDevices(ctx)
}

func (r *OpenVINORuntime) ListModels(ctx context.Context) ([]runtime.ModelInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries, err := os.ReadDir(r.modelPath)
	if err != nil {
		if r.core == nil {
			return []runtime.ModelInfo{}, nil
		}
		return nil, fmt.Errorf("cannot list models in %q: %w", r.modelPath, err)
	}

	dirs := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() {
			dirs[entry.Name()] = true
		}
	}

	models := make([]runtime.ModelInfo, 0, len(dirs))
	for dir := range dirs {
		info := runtime.ModelInfo{
			ID:   dir,
			Name: dir,
			Path: filepath.Join(r.modelPath, dir),
		}
		if lm, ok := r.models[dir]; ok {
			info.Loaded = true
			info.Precision = lm.info.Precision
			info.Size = lm.info.Size
		}
		models = append(models, info)
	}

	return models, nil
}

func (r *OpenVINORuntime) LoadModel(ctx context.Context, modelID, path, device string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.core == nil {
		slog.Warn("stub mode: cannot load model", "model_id", modelID)
		r.models[modelID] = &loadedModel{
			info: runtime.ModelInfo{ID: modelID, Name: modelID, Loaded: true},
		}
		return nil
	}

	if device == "" || device == "auto" {
		device = r.defaultDev
	}

	lm, exists := r.models[modelID]
	if exists {
		if _, ok := lm.compiledByDevice[device]; ok {
			slog.Debug("model already compiled for device", "model_id", modelID, "device", device)
			return nil
		}
	} else {
		modelPath := path
		if !filepath.IsAbs(path) {
			modelPath = filepath.Join(r.modelPath, path)
		}

		xmlPath := findModelFile(modelPath)
		if xmlPath == "" {
			return fmt.Errorf("model file not found in %q (expected .xml)", modelPath)
		}

		ovModel, err := r.core.ReadModel(xmlPath)
		if err != nil {
			return fmt.Errorf("failed to read model %q: %w", modelID, err)
		}

		compiled, err := r.core.CompileModel(ovModel, device)
		if err != nil {
			ovModel.Free()
			return fmt.Errorf("failed to compile model %q on %s: %w", modelID, device, err)
		}

		fi, _ := os.Stat(xmlPath)
		size := int64(0)
		if fi != nil {
			binPath := strings.TrimSuffix(xmlPath, ".xml") + ".bin"
			if bfi, err := os.Stat(binPath); err == nil {
				size = bfi.Size()
			}
		}

		lm = &loadedModel{
			info: runtime.ModelInfo{
				ID:     modelID,
				Name:   modelID,
				Path:   modelPath,
				Loaded: true,
				Size:   size,
			},
			model:            ovModel,
			compiledByDevice: make(map[string]*compiledModelHandle),
		}
		lm.compiledByDevice[device] = &compiledModelHandle{
			compiled: compiled,
			device:   device,
		}
		r.models[modelID] = lm
		slog.Info("model loaded", "model_id", modelID, "device", device, "size", size)
		return nil
	}

	compiled, err := r.core.CompileModel(lm.model, device)
	if err != nil {
		return fmt.Errorf("failed to compile model %q on %s: %w", modelID, device, err)
	}

	lm.compiledByDevice[device] = &compiledModelHandle{
		compiled: compiled,
		device:   device,
	}
	slog.Info("model compiled for additional device", "model_id", modelID, "device", device, "total_devices", len(lm.compiledByDevice))
	return nil
}

func findModelFile(path string) string {
	matches, err := filepath.Glob(filepath.Join(path, "*.xml"))
	if err != nil || len(matches) == 0 {
		if fi, err := os.Stat(path); err == nil && !fi.IsDir() && strings.HasSuffix(path, ".xml") {
			return path
		}
		return ""
	}
	if len(matches) == 1 {
		return matches[0]
	}
	for _, m := range matches {
		if strings.Contains(m, "openvino") {
			return m
		}
	}
	return matches[0]
}

func (r *OpenVINORuntime) UnloadModel(ctx context.Context, modelID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	lm, ok := r.models[modelID]
	if !ok {
		return fmt.Errorf("model %q is not loaded", modelID)
	}

	for _, ch := range lm.compiledByDevice {
		ch.compiled.Free()
	}
	if lm.model != nil {
		lm.model.Free()
	}
	delete(r.models, modelID)

	slog.Info("model unloaded", "model_id", modelID)
	return nil
}
