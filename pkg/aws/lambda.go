package aws

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/log"
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

func NewHandler(ctx context.Context, logger *slog.Logger, version string, logLevel *slog.LevelVar) (*Handler, error) {
	// read config from the environment variable
	// parse config as YAML
	cfg := &config.Config{}
	if err := config.Read(cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := log.SetLevel(logLevel, cfg.LogLevel); err != nil {
		return nil, fmt.Errorf("set log level: %w", err)
	}

	// read secrets from AWS SecretsManager
	secret, err := readSecret(ctx, cfg.AWS.SecretID)
	if err != nil {
		return nil, fmt.Errorf("get secret from AWS Secrets Manager: %w", err)
	}
	ctrl, err := controller.New(&controller.InputNew{
		Config:              cfg,
		Version:             version,
		WebhookSecret:       []byte(secret.WebhookSecret),
		GitHubAppPrivateKey: secret.GitHubAppPrivateKey,
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
