package openvino

import (
	"context"

	"github.com/openforge-ai/openforge/runtime"
)

type Provider struct {
	rt *OpenVINORuntime
}

func NewProvider(modelPath string) *Provider {
	return &Provider{
		rt: NewRuntime(modelPath),
	}
}

func (p *Provider) Name() string {
	return "openvino"
}

func (p *Provider) Runtime() runtime.Runtime {
	return p.rt
}

func (p *Provider) Initialize(ctx context.Context) error {
	return p.rt.Initialize(ctx)
}

func (p *Provider) Shutdown(ctx context.Context) error {
	return p.rt.Close(ctx)
}
