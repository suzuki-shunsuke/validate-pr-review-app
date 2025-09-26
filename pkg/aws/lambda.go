package aws

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
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

	// read secrets from AWS SecretsManager
	secret, err := readSecret(ctx, cfg.AWS.SecretID)
	if err != nil {
		return nil, fmt.Errorf("get secret from AWS Secrets Manager: %w", err)
	}
	ctrl, err := controller.New(&controller.InputNew{
		Config:        cfg,
		Version:       version,
		WebhookSecret: []byte(secret.WebhookSecret),
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
	var handler any
	if h.config.AWS.UseLambdaFunctionURL {
		handler = h.handleFunctionURL
	} else {
		handler = h.do
	}
	lambda.StartWithOptions(handler, lambda.WithContext(ctx))
}

func (h *Handler) newLogger(ctx context.Context) *slog.Logger {
	logger := h.logger
	lc, ok := lambdacontext.FromContext(ctx)
	if ok {
		return logger.With("aws_request_id", lc.AwsRequestID)
	}
	logger.Warn("lambda context is not found")
	return logger
}
