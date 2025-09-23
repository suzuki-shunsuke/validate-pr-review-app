package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/google/go-github/v74/github"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var (
	errHeaderXGitHubHookInstallationTargetIDIsRequired  = errors.New("header X-GITHUB-HOOK-INSTALLATION-TARGET-ID is required")
	errHeaderXGitHubHookInstallationTargetIDMustBeInt64 = errors.New("header X-GITHUB-HOOK-INSTALLATION-TARGET-ID must be integer")
	errHeaderXHubSignatureIsRequired                    = errors.New("header X-HUB-SIGNATURE is required")
	errSignatureInvalid                                 = errors.New("signature is invalid")
	errHeaderXHubEventIsRequired                        = errors.New("header X-HUB-EVENT is required")
	errInvalidEventType                                 = errors.New("event type is invalid")
	errInvalidAppID                                     = errors.New("app ID is invalid")
)

const (
	headerXGitHubHookInstallationTargetID = "X-GITHUB-HOOK-INSTALLATION-TARGET-ID"
	headerXHubSignature                   = "X-HUB-SIGNATURE"
	headerXGitHubEvent                    = "X-GITHUB-EVENT"
	eventPullRequestReview                = "pull_request_review"
)

func (h *Handler) validateRequest(logger *slog.Logger, req *Request) (*github.PullRequestReviewEvent, error) {
	headers := req.Params.Headers
	bodyStr := req.Body
	appIDstr, ok := headers[headerXGitHubHookInstallationTargetID]
	if !ok {
		return nil, errHeaderXGitHubHookInstallationTargetIDIsRequired
	}
	appID, err := strconv.ParseInt(appIDstr, 10, 64)
	if err != nil {
		return nil, errHeaderXGitHubHookInstallationTargetIDMustBeInt64
	}
	if appID != h.config.AppID {
		return nil, slogerr.With(errInvalidAppID, "app_id", appID, "expected_app_id", h.config.AppID) //nolint:wrapcheck
	}

	sig, ok := headers[headerXHubSignature]
	if !ok {
		return nil, errHeaderXHubSignatureIsRequired
	}

	bodyB := []byte(bodyStr)
	if err := github.ValidateSignature(sig, bodyB, h.webhookSecret); err != nil {
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
