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

// Handler represents the main AWS Lambda handler for processing GitHub webhook events.
// It manages the application configuration, secrets, and controller initialization.
type Handler struct {
	logger     *slog.Logger      // Structured logger for the handler
	config     *config.Config    // Application configuration
	controller Controller        // Controller interface for processing requests
}

// Controller defines the interface for processing webhook requests.
type Controller interface {
	Run(ctx context.Context, logger *slog.Logger, req *controller.Request) error
}

// NewHandler creates a new Lambda handler with the provided configuration.
// It reads configuration from environment variables, retrieves secrets from AWS Secrets Manager,
// and initializes the controller for processing GitHub webhook events.
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

	// Read secrets from AWS Secrets Manager
	secret, err := readSecret(ctx, cfg.AWS.SecretID)
	if err != nil {
		return nil, fmt.Errorf("get secret from AWS Secrets Manager: %w", err)
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

// Start begins the Lambda handler execution with the appropriate handler function.
// It selects between Function URL and API Gateway proxy handlers based on configuration.
func (h *Handler) Start(ctx context.Context) {
	lambda.StartWithOptions(h.handler(), lambda.WithContext(ctx))
}

// handler returns the appropriate handler function based on the AWS configuration.
// It returns either the Function URL handler or the API Gateway proxy handler.
func (h *Handler) handler() any {
	if h.config.AWS.UseLambdaFunctionURL {
		return h.handleFunctionURL
	}
	return h.handleProxy
}

// newLogger creates a context-aware logger that includes the AWS request ID for tracing.
// It extracts the request ID from the Lambda context and adds it to the logger.
func (h *Handler) newLogger(ctx context.Context) (*slog.Logger, string) {
	logger := h.logger
	lc, ok := lambdacontext.FromContext(ctx)
	if ok {
		return logger.With("aws_request_id", lc.AwsRequestID), lc.AwsRequestID
	}
	logger.Warn("lambda context is not found")
	return logger, ""
}
