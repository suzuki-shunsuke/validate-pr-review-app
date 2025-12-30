package aws

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/secret"
)

// Read config and secret
// Invoke validator
// Pass the webhook to validator

type Handler struct {
	logger     *slog.Logger
	config     *config.Config
	controller Controller
}

type Controller interface {
	Run(ctx context.Context, logger *slog.Logger, req *controller.Request) error
}

func NewHandler(logger *slog.Logger, version string, cfg *config.Config, s *secret.Secret) (*Handler, error) {
	ctrl, err := controller.New(&controller.InputNew{
		Config:              cfg,
		Version:             version,
		WebhookSecret:       []byte(s.WebhookSecret),
		GitHubAppPrivateKey: s.GitHubAppPrivateKey,
		Logger:              logger,
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

func (h *Handler) Start(ctx context.Context) {
	lambda.StartWithOptions(h.handler(), lambda.WithContext(ctx))
}

func (h *Handler) handler() any {
	if h.config.AWS.UseLambdaFunctionURL {
		return h.handleFunctionURL
	}
	return h.handleProxy
}

func (h *Handler) newLogger(ctx context.Context) (*slog.Logger, string) {
	logger := h.logger
	lc, ok := lambdacontext.FromContext(ctx)
	if ok {
		return logger.With("aws_request_id", lc.AwsRequestID), lc.AwsRequestID
	}
	logger.Warn("lambda context is not found")
	return logger, ""
}
