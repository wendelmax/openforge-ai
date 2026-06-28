package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/internal/engine"
	"github.com/openforge-ai/openforge/internal/provider/openvino"
	"github.com/openforge-ai/openforge/internal/server"
	"github.com/spf13/cobra"
)

var serveFlags struct {
	port    int
	model   string
	device  string
	modelPath string
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the OpenForge HTTP server",
	Long: `Start the OpenForge inference server.

The server exposes an OpenAI-compatible REST API at http://localhost:9090.
Models are loaded from the configured model directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}

		if serveFlags.port > 0 {
			cfg.Server.Port = serveFlags.port
		}
		if serveFlags.model != "" {
			cfg.Models.Default = serveFlags.model
		}
		if serveFlags.device != "" {
			cfg.Models.Device = serveFlags.device
		}
		if serveFlags.modelPath != "" {
			cfg.Models.Path = serveFlags.modelPath
		}

		logLevel := slog.LevelInfo
		if verbose {
			logLevel = slog.LevelDebug
		}
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		provider := openvino.NewProvider(cfg.Models.Path)

		slog.Info("initializing OpenVINO provider")
		if err := provider.Initialize(ctx); err != nil {
			return fmt.Errorf("provider init: %w", err)
		}

		if rt, ok := provider.Runtime().(*openvino.OpenVINORuntime); ok {
			rt.SetDeviceConfig(&cfg.Devices)
		}

		if cfg.Models.Default != "" {
			slog.Info("loading default model", "model", cfg.Models.Default, "device", cfg.Models.Device)
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

		eng := engine.New(provider.Runtime())
		srv := server.New(eng, cfg)

		go func() {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh
			slog.Info("shutting down gracefully...")
			cancel()
			srv.Shutdown(ctx)
			provider.Shutdown(ctx)
		}()

		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		slog.Info("server started", "address", addr, "device", cfg.Models.Device)
		return srv.Start(addr)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&serveFlags.port, "port", "p", 0, "server port")
	serveCmd.Flags().StringVarP(&serveFlags.model, "model", "m", "", "default model")
	serveCmd.Flags().StringVarP(&serveFlags.device, "device", "d", "", "inference device")
	serveCmd.Flags().StringVar(&serveFlags.modelPath, "models", "", "model directory")
}
