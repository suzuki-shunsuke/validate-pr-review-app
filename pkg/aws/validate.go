package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/google/go-github/v74/github"
)

var (
	errHeaderXGitHubHookInstallationTargetIDIsRequired   = errors.New("header X-GITHUB-HOOK-INSTALLATION-TARGET-ID is required")
	errHeaderXGitHubHookInstallationTargetIDMustBeInt64 = errors.New("header X-GITHUB-HOOK-INSTALLATION-TARGET-ID must be integer")
	errHeaderXHubSignatureIsRequired                    = errors.New("header X-HUB-SIGNATURE is required")
	errSignatureInvalid                                 = errors.New("signature is invalid")
	errHeaderXHubEventIsRequired                        = errors.New("header X-HUB-EVENT is required")
)

func (h *Handler) validate(logger *slog.Logger, req *Request) (*github.PullRequestReviewEvent, error) {
	headers := req.Params.Headers
	bodyStr := req.Body
	appIDstr, ok := headers["X-GITHUB-HOOK-INSTALLATION-TARGET-ID"]
	if !ok {
		return nil, errHeaderXGitHubHookInstallationTargetIDIsRequired
	}
	appID, err := strconv.ParseInt(appIDstr, 10, 64)
	if err != nil {
		return nil, errHeaderXGitHubHookInstallationTargetIDMustBeInt64
	}
	if appID != h.config.AppID {
		return nil, fmt.Errorf("app ID %d is not supported, expected %d", appID, h.config.AppID)
	}

	sig, ok := headers["X-HUB-SIGNATURE"]
	if !ok {
		return nil, errHeaderXHubSignatureIsRequired
	}

	bodyB := []byte(bodyStr)
	if err := github.ValidateSignature(sig, bodyB, h.webhookSecret); err != nil {
		logger.Warn("validate the webhook signature", "error", err)
		return nil, errSignatureInvalid
	}

	evType, ok := headers["X-GITHUB-EVENT"]
	if !ok {
		return nil, errHeaderXHubEventIsRequired
	}
	if evType != "pull_request_review" {
		return nil, fmt.Errorf("event type %q is not supported, expected %q", evType, "pull_request_review")
	}

	payload := &github.PullRequestReviewEvent{}
	if err := json.Unmarshal(bodyB, payload); err != nil {
		logger.Warn("parse a webhook payload", "error", err)
		return nil, fmt.Errorf("parse a webhook payload: %w", err)
	}

	return payload, nil
}
