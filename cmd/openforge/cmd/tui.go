package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/openforge-ai/openforge/internal/pm"
	"github.com/openforge-ai/openforge/internal/provider/openvino"
	"github.com/openforge-ai/openforge/internal/tool"
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

	ctx := context.Background()

	// Build the OpenVINO adapter + discover other providers
	ovProvider := pm.NewOpenVINOAdapter(cfg.ModelPath)
	if err := ovProvider.Start(ctx); err != nil {
		slog.Warn("OpenVINO init failed (stub mode)", "error", err)
	}

	pman := pm.New([]pm.Provider{ovProvider}, pm.Config{
		Chain: []pm.ProviderType{pm.ProviderOpenVINO, pm.ProviderOllama},
	})

	prov, err := pman.ActiveProvider(ctx)
	if err != nil {
		return fmt.Errorf("no inference provider available: %w\nRun 'openforge provider install ollama' to install one.", err)
	}

	tools := tool.DefaultRegistry().List()

	// Load default model if configured
	if cfg.DefaultModel != "" {
		if err := prov.LoadModel(ctx, cfg.DefaultModel); err != nil {
			slog.Warn("default model not loaded", "model", cfg.DefaultModel, "error", err)
		}
	}

	mdl := tui.New(pman, prov, tools)
	p := tea.NewProgram(mdl, tea.WithAltScreen())
	mdl.SetProgram(p)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	return nil
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
	return &tuiConfig{ModelPath: "./models", DefaultModel: "", Device: "auto"}
}

var _ = openvino.NewProvider // ensure import used
