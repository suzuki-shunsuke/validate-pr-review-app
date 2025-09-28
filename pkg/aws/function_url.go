package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
)

type FunctionURLRequest struct {
	payload any
	request *events.APIGatewayV2HTTPRequest
	err     error
}

func (p *FunctionURLRequest) Validate() error {
	if p.request == nil {
		return errors.New("request is nil")
	}
	if p.request.Body == "" {
		return errors.New("body is empty")
	}
	if p.request.Headers == nil {
		return errors.New("headers are empty")
	}
	return nil
}

func (p *FunctionURLRequest) UnmarshalJSON(b []byte) error {
	p.err = p.unmarshalJSON(b)
	return nil
}

func (p *FunctionURLRequest) unmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &p.request); err != nil {
		if err := json.Unmarshal(b, &p.payload); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		return fmt.Errorf("unmarshal request to APIGatewayV2HTTPRequest: %w", err)
	}
	if err := p.Validate(); err != nil {
		if err := json.Unmarshal(b, &p.payload); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		return err
	}
	return nil
}

func (h *Handler) handleFunctionURL(ctx context.Context, req *FunctionURLRequest) {
	logger := h.newLogger(ctx)
	if err := h.controller.Run(ctx, logger, &controller.Request{
		Body:    req.request.Body,
		Headers: req.request.Headers,
	}); err != nil {
		slogerr.WithError(logger, err).Error("handle request")
	}
}
