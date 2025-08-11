package aws

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
)

type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h *Handler) Start(ctx context.Context) error {
	// TODO read config from the environment variable
	// TODO parse config as YAML
	// TODO read secrets from AWS SecretsManager
	lambda.StartWithOptions(h.do, lambda.WithContext(ctx))
	return nil
}

func (h *Handler) do(_ context.Context, req any) error {
	h.logger.Info("Starting a request", "request", req)
	defer h.logger.Info("Ending a request", "request", req)
	// TODO parse webhook payload
	return nil
}
