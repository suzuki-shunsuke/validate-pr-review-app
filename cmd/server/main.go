package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/log"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/secret"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/server"
)

var version = ""

func main() {
	if code := run(); code != 0 {
		os.Exit(code)
	}
}

func run() int {
	logLevel := &slog.LevelVar{}
	logger := log.New(os.Stderr, version, logLevel)
	if err := core(logger, logLevel); err != nil {
		slogerr.WithError(logger, err).Error("failed")
		return 1
	}
	return 0
}

func core(logger *slog.Logger, logLevel *slog.LevelVar) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg := &config.Config{}
	if err := config.Read(cfg); err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	if err := log.SetLevel(logLevel, cfg.LogLevel); err != nil {
		return fmt.Errorf("set log level: %w", err)
	}
	s := &secret.Secret{}
	if err := secret.Read(s); err != nil {
		return fmt.Errorf("read secret: %w", err)
	}
	server, err := server.New(logger, version, cfg, s)
	if err != nil {
		return fmt.Errorf("create a new server: %w", err)
	}
	server.Start(ctx)
	return nil
}
