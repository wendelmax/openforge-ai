// Package runtime defines the Provider interface for runtime backend implementations.
package runtime

import "context"

// Provider defines the interface for a runtime backend that manages a Runtime instance.
type Provider interface {
	Name() string
	Runtime() Runtime
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
