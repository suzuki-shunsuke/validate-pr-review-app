package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/aws"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/log"
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
	handler, err := aws.NewHandler(ctx, logger, version, logLevel)
	if err != nil {
		return fmt.Errorf("create a new handler: %w", err)
	}
	handler.Start(ctx)
	return nil
}
