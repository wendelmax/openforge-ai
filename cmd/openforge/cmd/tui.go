package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/openforge-ai/openforge/internal/provider/openvino"
	"github.com/openforge-ai/openforge/internal/tui"
)

func runTUI() error {
	cfg, err := loadConfig()
	if err != nil {
		slog.Warn("no config file, using defaults")
		cfg = defaultConfig()
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))

	provider := openvino.NewProvider(cfg.ModelPath)

	ctx := context.Background()
	if err := provider.Initialize(ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	if cfg.DefaultModel != "" {
		if err := provider.Runtime().LoadModel(ctx, cfg.DefaultModel, cfg.DefaultModel, cfg.Device); err != nil {
			slog.Warn("default model not loaded", "model", cfg.DefaultModel, "error", err)
		}
	}

	mdl := tui.New(provider.Runtime())
	p := tea.NewProgram(mdl, tea.WithAltScreen())
	mdl.SetProgram(p)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	return provider.Shutdown(ctx)
}

type tuiConfig struct {
	ModelPath    string
	DefaultModel string
	Device       string
}

func loadConfig() (*tuiConfig, error) {
	if cfgFile == "" {
		return nil, fmt.Errorf("no config file")
	}
	return parseConfig(cfgFile)
}

func parseConfig(path string) (*tuiConfig, error) {
	return &tuiConfig{
		ModelPath:    "./models",
		DefaultModel: "",
		Device:       "auto",
	}, nil
}

func defaultConfig() *tuiConfig {
	return &tuiConfig{
		ModelPath:    "./models",
		DefaultModel: "",
		Device:       "auto",
	}
}
