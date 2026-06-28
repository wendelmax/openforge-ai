// Package config manages application configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ServerConfig represents the server configuration.
type ServerConfig struct {
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	Timeout int    `mapstructure:"timeout"`
}

// ModelsConfig represents the model configuration.
type ModelsConfig struct {
	Path    string `mapstructure:"path"`
	Default string `mapstructure:"default"`
	Device  string `mapstructure:"device"`
}

// CacheConfig represents the cache configuration.
type CacheConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Backend string `mapstructure:"backend"`
	Path    string `mapstructure:"path"`
	MaxSize int64  `mapstructure:"max_size"`
}

// DeviceConfig specifies device selection per workload, with a global default.
type DeviceConfig struct {
	Default   string `mapstructure:"default"`
	Chat      string `mapstructure:"chat"`
	Embedding string `mapstructure:"embedding"`
	Rerank    string `mapstructure:"rerank"`
}

// LoggingConfig represents the logging configuration.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// TelemetryConfig represents the telemetry configuration.
type TelemetryConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// BenchmarkConfig controls startup benchmark behavior.
type BenchmarkConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Iterations int    `mapstructure:"iterations"`
	Prompt     string `mapstructure:"prompt"`
	MaxTokens  int    `mapstructure:"max_tokens"`
}

// Config is the top-level application configuration.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Models    ModelsConfig    `mapstructure:"models"`
	Cache     CacheConfig     `mapstructure:"cache"`
	Devices   DeviceConfig    `mapstructure:"devices"`
	Benchmark BenchmarkConfig `mapstructure:"benchmark"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Telemetry TelemetryConfig `mapstructure:"telemetry"`
}

// Default returns a Config with sensible default values.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host:    "127.0.0.1",
			Port:    9090,
			Timeout: 30,
		},
		Models: ModelsConfig{
			Path:   "./models",
			Device: "auto",
		},
		Cache: CacheConfig{
			Enabled: true,
			Backend: "memory",
			Path:    "./data/cache",
			MaxSize: 1073741824,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Devices: DeviceConfig{
			Default:   "auto",
			Chat:      "",
			Embedding: "",
			Rerank:    "",
		},
		Benchmark: BenchmarkConfig{
			Enabled:    false,
			Iterations: 3,
			Prompt:     "The quick brown fox jumps over the lazy dog",
			MaxTokens:  50,
		},
		Telemetry: TelemetryConfig{
			Enabled: false,
		},
	}
}

// Load reads configuration from a file or searches standard paths if path is empty.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("server.port", 9090)
	v.SetDefault("server.timeout", 30)
	v.SetDefault("models.path", "./models")
	v.SetDefault("models.default", "")
	v.SetDefault("models.device", "auto")
	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.backend", "memory")
	v.SetDefault("cache.path", "./data/cache")
	v.SetDefault("cache.max_size", int64(1073741824))
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("devices.default", "auto")
	v.SetDefault("devices.chat", "")
	v.SetDefault("devices.embedding", "")
	v.SetDefault("devices.rerank", "")
	v.SetDefault("benchmark.enabled", false)
	v.SetDefault("benchmark.iterations", 3)
	v.SetDefault("benchmark.prompt", "The quick brown fox jumps over the lazy dog")
	v.SetDefault("benchmark.max_tokens", 50)
	v.SetDefault("telemetry.enabled", false)

	v.SetEnvPrefix("OPENFORGE")
	v.AutomaticEnv()

	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("config path: %w", err)
		}
		if _, err := os.Stat(absPath); err != nil {
			return nil, fmt.Errorf("config file %q: %w", absPath, err)
		}
		v.SetConfigFile(absPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config %q: %w", absPath, err)
		}
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("$HOME/.openforge")
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("read config: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
