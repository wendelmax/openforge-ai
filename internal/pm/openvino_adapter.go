package pm

import (
	"context"

	"github.com/openforge-ai/openforge/internal/provider/openvino"
	"github.com/openforge-ai/openforge/runtime"
)

// OpenVINOAdapter adapts the existing OpenVINO runtime to the pm.Provider interface.
type OpenVINOAdapter struct {
	rt   *openvino.OpenVINORuntime
	info ProviderInfo
}

func NewOpenVINOAdapter(modelPath string) *OpenVINOAdapter {
	rt := openvino.NewRuntime(modelPath)
	return &OpenVINOAdapter{
		rt: rt,
		info: ProviderInfo{
			Type:               ProviderOpenVINO,
			Name:               "OpenVINO",
			Description:        "Intel's native inference runtime for CPU, GPU, and NPU.",
			Website:            "https://docs.openvino.ai",
			SupportedHardware:  []string{"CPU", "GPU (Intel Iris/Arc)", "NPU (Intel AI Boost)"},
			SupportedWorkloads: []WorkloadType{WorkloadChat, WorkloadCode, WorkloadEmbed, WorkloadRerank},
			Native:             true,
			NeedsInstall:       false,
			AutoStartable:      false,
		},
	}
}

func (a *OpenVINOAdapter) Info() ProviderInfo { return a.info }

func (a *OpenVINOAdapter) Status(ctx context.Context) (*ProviderHealth, error) {
	devices, err := a.rt.ListDevices(ctx)
	if err != nil {
		return HealthError(err), nil
	}
	availCount := 0
	deviceNames := make([]string, 0)
	for _, d := range devices {
		if d.Available {
			availCount++
			deviceNames = append(deviceNames, string(d.Type))
		}
	}
	if availCount == 0 {
		return HealthUnavailable("no available devices"), nil
	}
	h := HealthAvailable()
	h.Devices = deviceNames
	return h, nil
}

func (a *OpenVINOAdapter) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	prompt := ""
	for _, m := range req.Messages {
		prompt += m.Role + ": " + m.Content + "\n"
	}
	params := runtime.InferenceParams{
		Temperature: req.Temperature,
		TopK:        req.TopK,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Device:      req.Device,
	}
	result, err := a.rt.Infer(ctx, req.Model, prompt, params)
	if err != nil {
		return nil, err
	}
	return &ChatResponse{
		Model:    req.Model,
		Content:  result.Text,
		Provider: ProviderOpenVINO,
		Device:   req.Device,
	}, nil
}

func (a *OpenVINOAdapter) ChatStream(ctx context.Context, req *ChatRequest) (<-chan Token, error) {
	prompt := ""
	for _, m := range req.Messages {
		prompt += m.Role + ": " + m.Content + "\n"
	}
	params := runtime.InferenceParams{
		Temperature: req.Temperature,
		TopK:        req.TopK,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Device:      req.Device,
	}
	streamCh, err := a.rt.InferStream(ctx, req.Model, prompt, params)
	if err != nil {
		return nil, err
	}
	ch := make(chan Token, 64)
	go func() {
		defer close(ch)
		for token := range streamCh {
			ch <- Token{Content: token}
		}
		ch <- Token{Done: true}
	}()
	return ch, nil
}

func (a *OpenVINOAdapter) Embed(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error) {
	result, err := a.rt.Embed(ctx, req.Model, req.Inputs, req.Device)
	if err != nil {
		return nil, err
	}
	return &EmbedResponse{
		Model:      req.Model,
		Embeddings: result.Embeddings,
		Provider:   ProviderOpenVINO,
	}, nil
}

func (a *OpenVINOAdapter) ListModels(ctx context.Context) ([]Model, error) {
	modelList, err := a.rt.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	models := make([]Model, len(modelList))
	for i, m := range modelList {
		models[i] = Model{
			ID:       m.ID,
			Name:     m.Name,
			Provider: ProviderOpenVINO,
			Size:     m.Size,
			Loaded:   m.Loaded,
		}
	}
	return models, nil
}

func (a *OpenVINOAdapter) LoadModel(ctx context.Context, modelID string) error {
	return a.rt.LoadModel(ctx, modelID, modelID, "")
}

func (a *OpenVINOAdapter) UnloadModel(ctx context.Context, modelID string) error {
	return a.rt.UnloadModel(ctx, modelID)
}

func (a *OpenVINOAdapter) Start(ctx context.Context) error { return a.rt.Initialize(ctx) }

func (a *OpenVINOAdapter) Stop(ctx context.Context) error { return a.rt.Close(ctx) }
