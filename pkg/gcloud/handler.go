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

// Handler represents the main Google Cloud Functions handler for processing GitHub webhook events.
// It manages the application configuration, secrets, and controller initialization for Cloud Functions runtime.
type Handler struct {
	logger     *slog.Logger   // Structured logger for the handler
	config     *config.Config // Application configuration
	controller Controller     // Controller interface for processing requests
}

// Controller defines the interface for processing webhook requests in the Google Cloud Functions environment.
type Controller interface {
	Run(ctx context.Context, logger *slog.Logger, req *controller.Request) error
}

// NewHandler creates a new Google Cloud Functions handler with the provided configuration.
// It reads configuration from environment variables, retrieves secrets from Google Cloud Secret Manager,
// and initializes the controller for processing GitHub webhook events in Cloud Functions.
func NewHandler(ctx context.Context, logger *slog.Logger, version string, logLevel *slog.LevelVar) (*Handler, error) {
	// Read configuration from the CONFIG environment variable
	cfg := &config.Config{}
	if err := readConfig(cfg); err != nil {
		return nil, err
	}

	// Update log level if specified in configuration
	if cfg.LogLevel != "" {
		lvl, err := log.ParseLevel(cfg.LogLevel)
		if err != nil {
			return nil, fmt.Errorf("parse log level: %w", slogerr.With(err, "log_level", cfg.LogLevel))
		}
		logLevel.Set(lvl)
	}

	// Read secrets from Google Cloud Secret Manager
	secret, err := readSecret(ctx, cfg.GoogleCloud.SecretName)
	if err != nil {
		return nil, fmt.Errorf("get secret from Google Cloud SecretManager: %w", err)
	}
	if err := secret.Validate(); err != nil {
		return nil, fmt.Errorf("validate secret: %w", err)
	}

	// Initialize the controller with configuration and secrets
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
