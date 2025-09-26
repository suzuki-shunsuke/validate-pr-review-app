package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/go-github/v75/github"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var (
	errHeaderXHubSignatureIsRequired = errors.New("header X-HUB-SIGNATURE is required")
	errSignatureInvalid              = errors.New("signature is invalid")
	errHeaderXHubEventIsRequired     = errors.New("header X-HUB-EVENT is required")
	errInvalidEventType              = errors.New("event type is invalid")
)

const (
	headerXGitHubHookInstallationTargetID = "X-GITHUB-HOOK-INSTALLATION-TARGET-ID"
	headerXHubSignature                   = "X-HUB-SIGNATURE"
	headerXGitHubEvent                    = "X-GITHUB-EVENT"
	eventPullRequestReview                = "pull_request_review"
)

func (c *Controller) validateRequest(logger *slog.Logger, req *Request) (*github.PullRequestReviewEvent, error) {
	headers := make(map[string]string, len(req.Params.Headers))
	for k, v := range req.Params.Headers {
		headers[strings.ToUpper(k)] = v
	}
	bodyStr := req.Body

	sig, ok := headers[headerXHubSignature]
	if !ok {
		return nil, errHeaderXHubSignatureIsRequired
	}

	bodyB := []byte(bodyStr)
	if err := github.ValidateSignature(sig, bodyB, c.input.WebhookSecret); err != nil {
		logger.Warn("validate the webhook signature", "error", err)
		return nil, errSignatureInvalid
	}

	evType, ok := headers[headerXGitHubEvent]
	if !ok {
		return nil, errHeaderXHubEventIsRequired
	}
	if evType != eventPullRequestReview {
		return nil, slogerr.With(errInvalidEventType, "event_type", evType) //nolint:wrapcheck
	}

	payload := &github.PullRequestReviewEvent{}
	if err := json.Unmarshal(bodyB, payload); err != nil {
		logger.Warn("parse a webhook payload", "error", err)
		return nil, fmt.Errorf("parse a webhook payload: %w", err)
	}

	return payload, nil
}
