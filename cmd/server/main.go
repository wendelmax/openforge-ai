package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/internal/engine"
	"github.com/openforge-ai/openforge/internal/mcp"
	"github.com/openforge-ai/openforge/internal/permission"
	"github.com/openforge-ai/openforge/internal/pm"
	"github.com/openforge-ai/openforge/internal/provider/openvino"
	"github.com/openforge-ai/openforge/internal/server"
)

var (
	configPath = flag.String("config", "config.yaml", "path to configuration file")
	port       = flag.Int("port", 0, "server port (overrides config)")
	verbose    = flag.Bool("verbose", false, "enable debug logging")
)

func main() {
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if *port > 0 {
		cfg.Server.Port = *port
	}

	if *verbose {
		cfg.Logging.Level = "debug"
	}

	logLevel := slog.LevelInfo
	if cfg.Logging.Level == "debug" {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Provider
	provider := openvino.NewProvider(cfg.Models.Path)
	rt := provider.Runtime()
	_ = engine.New(rt) // keep engine for legacy compat

	// MCP servers
	mcpRegistry := mcp.NewRegistry()
	mcpServers := config.GetMCPServers(cfg)
	if len(mcpServers) > 0 {
		slog.Info("connecting MCP servers", "count", len(mcpServers))
		for name, srv := range mcpServers {
			if err := mcpRegistry.Connect(ctx, name, srv.Command, srv.Args...); err != nil {
				slog.Error("MCP connection failed", "server", name, "error", err)
			}
		}
	}

	// Permissions
	permStore := permission.NewMemoryStore()
	permManager := mcp.BuildPermsFromConfig(cfg, permStore, "server")
	_ = permManager

	// Build PM adapter for server
	ovAdapter := pm.NewOpenVINOAdapter(cfg.Models.Path)
	pman := pm.New([]pm.Provider{ovAdapter}, pm.Config{
		Chain: []pm.ProviderType{pm.ProviderOpenVINO, pm.ProviderOllama},
	})

	_, err = pman.ActiveProvider(ctx)
	if err != nil {
		slog.Error("no provider available", "error", err)
		os.Exit(1)
	}

	eng := engine.New(rt)
	srv := server.New(eng, cfg)

	cleanup := func() {
		cancel()
		mcpRegistry.CloseAll()
		srv.Shutdown(ctx)
		provider.Shutdown(ctx)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	if err := srv.RunUntilSignal(ctx, addr, cleanup); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
