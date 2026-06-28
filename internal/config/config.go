// Package config manages application configuration.
package config

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// SessionConfig configures session persistence.
type SessionConfig struct {
	Backend string `mapstructure:"backend"`
	Path    string `mapstructure:"path"`
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
	Session   SessionConfig   `mapstructure:"session"`
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
		Session: SessionConfig{
			Backend: "memory",
			Path:    "./data/sessions",
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
	v.SetDefault("session.backend", "memory")
	v.SetDefault("session.path", "./data/sessions")
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

	expandConfig(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

// expandConfig walks all viper string values and expands env vars and shell commands.
// Supports: $VAR, ${VAR}, ${VAR:-default}, $(command).
// Config files are trusted code — $(command) runs arbitrary shell commands.
func expandConfig(v *viper.Viper) {
	for _, key := range v.AllKeys() {
		val := v.Get(key)
		if str, ok := val.(string); ok && containsDollar(str) {
			v.Set(key, expandValue(str))
		}
	}
}

func containsDollar(s string) bool {
	return strings.Contains(s, "$")
}

func expandValue(s string) string {
	s = expandCmdSubst(s)
	s = expandDollarVars(s)
	return s
}

// expandDollarVars replaces $VAR, ${VAR}, and ${VAR:-default} patterns.
// $$ is left as a literal $ (unlike os.Expand which treats $$ as $).
func expandDollarVars(s string) string {
	var buf bytes.Buffer
	for i := 0; i < len(s); i++ {
		if s[i] == '$' && i+1 < len(s) {
			if s[i+1] == '$' {
				buf.WriteByte('$')
				i++
				continue
			}
			if s[i+1] == '{' {
				j := strings.IndexByte(s[i+2:], '}')
				if j < 0 {
					buf.WriteByte(s[i])
					continue
				}
				name := s[i+2 : i+2+j]
				buf.WriteString(envVarMapping(name))
				i += 2 + j
				continue
			}
			// $NAME (simple env var)
			j := i + 1
			for j < len(s) && isEnvNameChar(s[j]) {
				j++
			}
			if j > i+1 {
				buf.WriteString(os.Getenv(s[i+1 : j]))
				i = j - 1
				continue
			}
		}
		buf.WriteByte(s[i])
	}
	return buf.String()
}

func isEnvNameChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func envVarMapping(name string) string {
	if idx := strings.Index(name, ":-"); idx >= 0 {
		varName := name[:idx]
		defaultVal := name[idx+2:]
		if v := os.Getenv(varName); v != "" {
			return v
		}
		return defaultVal
	}
	return os.Getenv(name)
}

func expandCmdSubst(s string) string {
	var buf bytes.Buffer
	for i := 0; i < len(s); i++ {
		if s[i] == '$' && i+2 < len(s) && s[i+1] == '(' {
			j := strings.IndexByte(s[i+2:], ')')
			if j < 0 {
				buf.WriteByte(s[i])
				continue
			}
			cmd := s[i+2 : i+2+j]
			out, err := exec.Command("sh", "-c", cmd).Output()
			if err == nil {
				buf.WriteString(strings.TrimRight(string(out), "\n\r"))
			}
			i += 2 + j
		} else {
			buf.WriteByte(s[i])
		}
	}
	return buf.String()
}
