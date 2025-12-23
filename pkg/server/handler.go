package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/secret"
)

type Server struct {
	logger     *slog.Logger
	config     *config.Config
	controller Controller
}

type Controller interface {
	Run(ctx context.Context, logger *slog.Logger, req *controller.Request) error
}

func New(logger *slog.Logger, version string, cfg *config.Config, secret *secret.Secret) (*Server, error) {
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
	return &Server{
		logger:     logger,
		config:     cfg,
		controller: ctrl,
	}, nil
}
