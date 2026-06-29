package lsp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoDetect_NoMarkers(t *testing.T) {
	m := NewManager()
	err := m.AutoDetect(t.TempDir())
	assert.NoError(t, err)
}

func TestManager_StartStopNonexistent(t *testing.T) {
	m := NewManager()
	err := m.StartLSP(context.Background(), "nonexistent", LSPConfig{
		Command: []string{"nonexistent-binary-xyz"},
	})
	assert.Error(t, err)
	assert.NoError(t, m.StopLSP(context.Background(), "nonexistent"))
}
