package aws

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
)

func (h *Handler) handleFunctionURL(ctx context.Context, req *events.APIGatewayV2HTTPRequest) {
	logger := h.newLogger(ctx)
	if err := h.controller.Run(ctx, logger, &controller.Request{
		Body: req.Body,
		Params: &controller.RequestParamsField{
			Headers: req.Headers,
		},
	}); err != nil {
		slogerr.WithError(logger, err).Error("handle request")
	}
}
