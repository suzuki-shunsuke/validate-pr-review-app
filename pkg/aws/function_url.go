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

// FunctionURLRequest represents a request received through AWS Lambda Function URLs.
// It handles both valid APIGatewayV2HTTPRequest events and invalid payloads for debugging.
type FunctionURLRequest struct {
	payload any                              // Stores the raw payload if parsing fails
	request *events.APIGatewayV2HTTPRequest  // The parsed Lambda Function URL request
	err     error                            // Error encountered during parsing
}

// Validate checks if the FunctionURLRequest contains all required fields.
// It ensures the request has a body and headers for processing GitHub webhooks.
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

// UnmarshalJSON implements json.Unmarshaler to parse Lambda Function URL events.
// It stores any parsing errors in the err field for later handling.
func (p *FunctionURLRequest) UnmarshalJSON(b []byte) error {
	p.err = p.unmarshalJSON(b)
	return nil
}

// unmarshalJSON attempts to parse the JSON as an APIGatewayV2HTTPRequest.
// If that fails, it falls back to storing the raw payload for debugging.
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

// handleFunctionURL processes Lambda Function URL requests.
// It creates a logger, validates the request, and forwards it to the controller.
func (h *Handler) handleFunctionURL(ctx context.Context, req *FunctionURLRequest) {
	logger, requestID := h.newLogger(ctx)
	if req.err != nil {
		if req.payload == nil {
			slogerr.WithError(logger, req.err).Warn("invalid request")
		} else {
			slogerr.WithError(logger, req.err).Warn("invalid request", "payload", req.payload)
		}
		return
	}

	if err := h.controller.Run(ctx, logger, &controller.Request{
		Body:      req.request.Body,
		Headers:   req.request.Headers,
		RequestID: requestID,
	}); err != nil {
		slogerr.WithError(logger, err).Error("handle request")
	}
}
