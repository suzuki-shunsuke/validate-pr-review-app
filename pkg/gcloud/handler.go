package gcloud

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/log"
)

type Handler struct {
	logger     *slog.Logger
	config     *config.Config
	controller Controller
}

type Controller interface {
	Run(ctx context.Context, logger *slog.Logger, req *controller.Request) error
}

func NewHandler(ctx context.Context, logger *slog.Logger, version string, logLevel *slog.LevelVar) (*Handler, error) {
	// read config from the environment variable
	// parse config as YAML
	cfg := &config.Config{}
	if err := readConfig(cfg); err != nil {
		return nil, err
	}

	if cfg.LogLevel != "" {
		lvl, err := log.ParseLevel(cfg.LogLevel)
		if err != nil {
			return nil, fmt.Errorf("parse log level: %w", slogerr.With(err, "log_level", cfg.LogLevel))
		}
		logLevel.Set(lvl)
	}

	// read secrets from Cloud SecretManager
	secret, err := readSecret(ctx, cfg.GoogleCloud.SecretName)
	if err != nil {
		return nil, fmt.Errorf("get secret from Google Cloud SecretManager: %w", err)
	}
	if err := secret.Validate(); err != nil {
		return nil, fmt.Errorf("validate secret: %w", err)
	}
	ctrl, err := controller.New(&controller.InputNew{
		Config:              cfg,
		Version:             version,
		WebhookSecret:       []byte(secret.WebhookSecret),
		GitHubAppPrivateKey: secret.GitHubAppPrivateKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create controller: %w", err)
	}
	return &Handler{
		logger:     logger,
		config:     cfg,
		controller: ctrl,
	}, nil
}
