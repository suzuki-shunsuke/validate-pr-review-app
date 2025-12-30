package aws

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
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

func NewHandler(logger *slog.Logger, ctrl Controller, cfg *config.Config) (*Handler, error) {
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
