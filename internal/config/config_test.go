package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, 30, cfg.Server.Timeout)
	assert.Equal(t, "./models", cfg.Models.Path)
	assert.Equal(t, "auto", cfg.Models.Device)
	assert.True(t, cfg.Cache.Enabled)
	assert.Equal(t, "memory", cfg.Cache.Backend)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.False(t, cfg.Telemetry.Enabled)
}

func TestDefaultServerConfig(t *testing.T) {
	cfg := Default()
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, 30, cfg.Server.Timeout)
}

func TestDefaultCacheConfig(t *testing.T) {
	cfg := Default()
	assert.True(t, cfg.Cache.Enabled)
	assert.Equal(t, "memory", cfg.Cache.Backend)
	assert.Equal(t, int64(1073741824), cfg.Cache.MaxSize)
}

func TestDefaultLoggingConfig(t *testing.T) {
	cfg := Default()
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
}

func TestDefaultTelemetryConfig(t *testing.T) {
	cfg := Default()
	assert.False(t, cfg.Telemetry.Enabled)
}

func TestLoad_NoPath_UsesDefaults(t *testing.T) {
	cfg, err := Load("")
	assert.NoError(t, err)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "auto", cfg.Models.Device)
	assert.True(t, cfg.Cache.Enabled)
}

func TestLoad_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.yaml"
	content := []byte(`
server:
  host: "0.0.0.0"
  port: 8080
models:
  device: "gpu"
  default: "phi-3-mini"
`)
	err := os.WriteFile(cfgPath, content, 0644)
	require.NoError(t, err)

	cfg, err := Load(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "gpu", cfg.Models.Device)
	assert.Equal(t, "phi-3-mini", cfg.Models.Default)
}

func TestLoad_InvalidPath(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config file")
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/bad.yaml"
	err := os.WriteFile(cfgPath, []byte("invalid: [yaml: broken"), 0644)
	require.NoError(t, err)

	_, err = Load(cfgPath)
	assert.Error(t, err)
}

func TestLoad_PartialConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/partial.yaml"
	content := []byte(`
cache:
  backend: "sqlite"
  path: "/tmp/cache"
`)
	err := os.WriteFile(cfgPath, content, 0644)
	require.NoError(t, err)

	cfg, err := Load(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, "sqlite", cfg.Cache.Backend)
	assert.Equal(t, "/tmp/cache", cfg.Cache.Path)
}

func TestDefault_ModelsConfig(t *testing.T) {
	cfg := Default()
	assert.Equal(t, "./models", cfg.Models.Path)
	assert.Equal(t, "auto", cfg.Models.Device)
	assert.Empty(t, cfg.Models.Default)
}

func TestDefault_CacheConfig(t *testing.T) {
	cfg := Default()
	assert.Equal(t, "./data/cache", cfg.Cache.Path)
	assert.Equal(t, int64(1073741824), cfg.Cache.MaxSize)
}

func TestDeviceConfigDefaults(t *testing.T) {
	cfg := Default()
	if cfg.Devices.Default != "auto" {
		t.Errorf("expected auto, got %q", cfg.Devices.Default)
	}
}
