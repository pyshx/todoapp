package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/pyshx/todoapp/internal/config"
	"github.com/pyshx/todoapp/internal/di"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("starting server",
		"version", cfg.Version,
		"go_version", runtime.Version(),
	)

	ctx := context.Background()
	container, err := di.New(ctx, cfg.DatabaseURL, cfg.GRPCPort, cfg.JWTSecret, cfg.JWTDuration, logger)
	if err != nil {
		logger.Error("failed to initialize dependencies", "error", err)
		os.Exit(1)
	}
	defer container.Close()

	logger.Info("connected to database")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("server ready", "port", cfg.GRPCPort)
		if err := container.Server.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-sigCh
	logger.Info("received shutdown signal", "signal", sig.String())

	if err := container.Server.GracefulShutdown(cfg.ShutdownTimeout); err != nil {
		logger.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
