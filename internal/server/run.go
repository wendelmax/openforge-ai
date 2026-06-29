package server

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// RunUntilSignal starts the server and blocks until a termination signal (SIGINT/SIGTERM)
// is received, then calls cleanup. Returns Start's error.
func (s *Server) RunUntilSignal(ctx context.Context, addr string, cleanup func()) error {
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down gracefully...")
		if cleanup != nil {
			cleanup()
		}
	}()

	slog.Info("server started", "address", addr)
	return s.Start(addr)
}
