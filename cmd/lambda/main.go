package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/aws"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/log"
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
	handler := aws.NewHandler(logger)
	return handler.Start(ctx) //nolint:wrapcheck
}
