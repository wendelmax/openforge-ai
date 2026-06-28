package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/internal/engine"
	"github.com/openforge-ai/openforge/internal/mcp"
	"github.com/openforge-ai/openforge/internal/permission"
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

	// --- Provider ---
	provider := openvino.NewProvider(cfg.Models.Path)
	if err := provider.Initialize(ctx); err != nil {
		slog.Error("provider initialization failed", "error", err)
		os.Exit(1)
	}
	if rt, ok := provider.Runtime().(*openvino.OpenVINORuntime); ok {
		rt.SetDeviceConfig(&cfg.Devices)
	}
	if cfg.Models.Default != "" {
		if err := provider.Runtime().LoadModel(ctx, cfg.Models.Default, cfg.Models.Default, cfg.Models.Device); err != nil {
			slog.Warn("default model not loaded", "model", cfg.Models.Default, "error", err)
		}
	}
	if cfg.Benchmark.Enabled && cfg.Models.Default != "" {
		if results, err := provider.Runtime().Benchmark(ctx, cfg.Models.Default, cfg.Benchmark.Iterations, cfg.Benchmark.Prompt, cfg.Benchmark.MaxTokens); err != nil {
			slog.Warn("benchmark failed", "error", err)
		} else {
			for device, res := range results {
				slog.Info("benchmark result", "device", device, "tokens_per_sec", res.ChatTokensPerSec)
			}
		}
	}

	// --- Session store ---
	sessionStore, err := engine.NewStoreFromConfig(cfg)
	if err != nil {
		slog.Error("failed to create session store", "error", err)
		os.Exit(1)
	}
	eng := engine.NewWithStore(provider.Runtime(), sessionStore)

	// --- MCP servers ---
	mcpRegistry := mcp.NewRegistry()
	if len(cfg.MCP) > 0 {
		slog.Info("connecting MCP servers", "count", len(cfg.MCP))
		if err := mcp.ConnectFromConfig(ctx, mcpRegistry, cfg); err != nil {
			slog.Error("MCP connection failed", "error", err)
		}
	}

	// --- Permissions ---
	grantsPath := filepath.Join(cfg.Session.Path, "grants.json")
	var permStore permission.Store
	permStore, err = permission.NewFileStore(grantsPath)
	if err != nil {
		slog.Warn("permission store init", "error", err)
		permStore = permission.NewMemoryStore()
	}
	permManager := mcp.BuildPermsFromConfig(cfg, permStore, "server")

	// --- MCP tool registration with permissions ---
	if len(mcpRegistry.List()) > 0 {
		skillExecutor := eng.SkillExecutor()
		if err := mcpRegistry.RegisterToolsWithPerms(ctx, skillExecutor, permManager); err != nil {
			slog.Error("MCP tool registration failed", "error", err)
		}
	}

	// --- Server ---
	srv := server.New(eng, cfg)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down gracefully...")
		cancel()
		mcpRegistry.CloseAll()
		srv.Shutdown(ctx)
		if err := provider.Shutdown(ctx); err != nil {
			slog.Error("provider shutdown error", "error", err)
		}
	}()

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	slog.Info("server started", "address", addr, "device", cfg.Models.Device)

	if err := srv.Start(addr); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
