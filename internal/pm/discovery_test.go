package pm

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscover_ReturnsAllFiveProviders(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	results, err := Discover(ctx)
	require.NoError(t, err)
	require.Len(t, results, 5)
	providers := make(map[ProviderType]bool)
	for _, r := range results { providers[r.Provider] = true }
	assert.True(t, providers[ProviderOpenVINO])
	assert.True(t, providers[ProviderOllama])
	assert.True(t, providers[ProviderLlamaCpp])
}

func TestDetectOpenVINO_FindsLibrary(t *testing.T) {
	libName := openvinoLibName()
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, libName), []byte("fake"), 0644)
	t.Setenv("OPENVINO_HOME", tmpDir)
	result := detectOpenVINO()
	assert.True(t, result.Installed)
	assert.True(t, result.Running)
}

func TestDetectOllama_HealthCheck(t *testing.T) {
	if testing.Short() { t.Skip("integration test") }
	if !binaryOnPath("ollama") { t.Skip("ollama not found") }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := detectOllama(ctx)
	assert.True(t, result.Installed)
}
