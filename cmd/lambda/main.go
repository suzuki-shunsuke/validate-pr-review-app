package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/aws"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/log"
)

var version = ""

func main() {
	logger := log.New(os.Stderr, version)
	if err := run(logger); err != nil {
		logger.Error("failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	logger.Info("Starting the application")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	handler, err := aws.NewHandler(ctx, logger)
	if err != nil {
		return fmt.Errorf("create a new handler: %w", err)
	}
	handler.Start(ctx)
	return nil
}
