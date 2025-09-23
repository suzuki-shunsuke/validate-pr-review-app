package aws

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (h *Handler) handleFunctionURL(ctx context.Context, req *events.APIGatewayV2HTTPRequest) {
	logger := h.logger
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		logger.Warn("lambda context is not found")
	} else {
		logger = logger.With("aws_request_id", lc.AwsRequestID)
	}
	if err := h.handle(ctx, logger, &Request{
		Body: req.Body,
		Params: &RequestParamsField{
			Headers: req.Headers,
		},
	}); err != nil {
		slogerr.WithError(logger, err).Error("handle request")
	}
}
