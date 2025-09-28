package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
)

// ProxyRequest represents a request received through AWS API Gateway proxy integration.
// It handles both valid APIGatewayProxyRequest events and invalid payloads for debugging.
type ProxyRequest struct {
	payload any                           // Stores the raw payload if parsing fails
	request *events.APIGatewayProxyRequest // The parsed API Gateway proxy request
	err     error                         // Error encountered during parsing
}

// Validate checks if the ProxyRequest contains all required fields.
// It ensures the request has a body and headers for processing GitHub webhooks.
func (p *ProxyRequest) Validate() error {
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

// UnmarshalJSON implements json.Unmarshaler to parse API Gateway proxy events.
// It stores any parsing errors in the err field for later handling.
func (p *ProxyRequest) UnmarshalJSON(b []byte) error {
	p.err = p.unmarshalJSON(b)
	return nil
}

// unmarshalJSON attempts to parse the JSON as an APIGatewayProxyRequest.
// If that fails, it falls back to storing the raw payload for debugging.
func (p *ProxyRequest) unmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &p.request); err != nil {
		if err := json.Unmarshal(b, &p.payload); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		return fmt.Errorf("unmarshal request to APIGatewayProxyRequest: %w", err)
	}
	if err := p.Validate(); err != nil {
		if err := json.Unmarshal(b, &p.payload); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		return err
	}
	return nil
}

// handleProxy processes API Gateway proxy requests and returns appropriate responses.
// It creates a logger, validates the request, forwards it to the controller, and returns an HTTP response.
func (h *Handler) handleProxy(ctx context.Context, req *ProxyRequest) (*events.APIGatewayProxyResponse, error) {
	logger, requestID := h.newLogger(ctx)
	if req.err != nil {
		if req.payload == nil {
			slogerr.WithError(logger, req.err).Warn("invalid request")
		} else {
			slogerr.WithError(logger, req.err).Warn("invalid request", "payload", req.payload)
		}
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       "OK",
		}, nil
	}
	if err := h.controller.Run(ctx, logger, &controller.Request{
		Body:      req.request.Body,
		Headers:   req.request.Headers,
		RequestID: requestID,
	}); err != nil {
		slogerr.WithError(logger, err).Error("handle request")
	}
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "OK",
	}, nil
}
