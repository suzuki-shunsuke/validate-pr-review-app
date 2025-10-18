package gcloud

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/log"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/server"
)

func New(ctx context.Context, logger *slog.Logger, version string, logLevel *slog.LevelVar) (*server.Server, error) {
	cfg := &config.Config{}
	if err := config.Read(cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := log.SetLevel(logLevel, cfg.LogLevel); err != nil {
		return nil, fmt.Errorf("set log level: %w", err)
	}

	secret, err := readSecret(ctx, cfg.GoogleCloud.SecretName)
	if err != nil {
		return nil, fmt.Errorf("get secret from Google Cloud SecretManager: %w", err)
	}
	srv, err := server.New(logger, version, cfg, secret)
	if err != nil {
		return nil, fmt.Errorf("create server: %w", err)
	}
	return srv, nil
}
