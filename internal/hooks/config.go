package hooks

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type HooksConfig struct {
	PreToolUse  []HookConfig `yaml:"PreToolUse"`
	PostToolUse []HookConfig `yaml:"PostToolUse"`
}

func LoadFromFile(path string) (*Engine, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read hooks config %q: %w", path, err)
	}
	var cfg HooksConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse hooks config %q: %w", path, err)
	}
	return LoadFromConfig(&cfg), nil
}

func LoadFromConfig(cfg *HooksConfig) *Engine {
	hooks := make([]HookConfig, 0, len(cfg.PreToolUse)+len(cfg.PostToolUse))
	for _, h := range cfg.PreToolUse {
		h.Events = []HookEvent{PreToolUse}
		hooks = append(hooks, h)
	}
	for _, h := range cfg.PostToolUse {
		h.Events = []HookEvent{PostToolUse}
		hooks = append(hooks, h)
	}
	return New(hooks)
}
