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

type ProxyRequest struct {
	payload any
	request *events.APIGatewayProxyRequest
	err     error
}

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

func (p *ProxyRequest) UnmarshalJSON(b []byte) error {
	p.err = p.unmarshalJSON(b)
	return nil
}

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
