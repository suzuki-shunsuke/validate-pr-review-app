package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"strconv"
	"strings"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
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
	eventMergeGroup                       = "merge_group"
)

func (c *Controller) verifySignature(body []byte, headers map[string]string) error {
	sig, ok := headers[headerXHubSignature]
	if !ok {
		return errHeaderXHubSignatureIsRequired
	}
	return c.validateSignature(sig, body, c.input.WebhookSecret)
}

func (c *Controller) normalizeHeaders(headers map[string]string) map[string]string {
	hs := make(map[string]string, len(headers))
	for k, v := range headers {
		hs[strings.ToUpper(k)] = v
	}
	return hs
}

func (c *Controller) verifyWebhook(logger *slog.Logger, req *Request) (*Event, error) {
	headers := c.normalizeHeaders(req.Params.Headers)
	body := []byte(req.Body)
	if err := c.verifySignature(body, headers); err != nil {
		slogerr.WithError(logger, err).Warn("validate the webhook signature")
		return nil, fmt.Errorf("validate the webhook signature: %w", err)
	}

	evType, ok := headers[headerXGitHubEvent]
	if !ok {
		return nil, errHeaderXHubEventIsRequired
	}
	switch evType {
	case eventPullRequestReview:
		payload := &github.PullRequestReviewEvent{}
		if err := json.Unmarshal(body, payload); err != nil {
			logger.Warn("parse a webhook payload", "error", err)
			return nil, fmt.Errorf("parse a webhook payload: %w", err)
		}
		return newPullRequestReviewEvent(payload), nil
	case eventMergeGroup:
		payload := &github.MergeGroupEvent{}
		if err := json.Unmarshal(body, payload); err != nil {
			logger.Warn("parse a webhook payload", "error", err)
			return nil, fmt.Errorf("parse a webhook payload: %w", err)
		}
		return newMergeGroupEvent(logger, payload)
	default:
		return nil, slogerr.With(errInvalidEventType, "event_type", evType) //nolint:wrapcheck
	}
}

type Event struct {
	Action       string
	RepoFullName string
	RepoOwner    string
	RepoName     string
	PRNumber     int
	ReviewState  string
	RepoID       string
	HeadSHA      string
}

func newPullRequestReviewEvent(ev *github.PullRequestReviewEvent) *Event {
	return &Event{
		Action:       ev.GetAction(),
		RepoFullName: ev.GetRepo().GetFullName(),
		RepoOwner:    ev.GetRepo().GetOwner().GetLogin(),
		RepoName:     ev.GetRepo().GetName(),
		PRNumber:     ev.GetPullRequest().GetNumber(),
		ReviewState:  ev.GetReview().GetState(),
		RepoID:       ev.GetRepo().GetNodeID(),
		HeadSHA:      ev.GetPullRequest().GetHead().GetSHA(),
	}
}

func getPRNumberFromPushBranch(logger *slog.Logger, ref string) (int, error) {
	branch, ok := strings.CutPrefix(ref, "refs/heads/")
	if !ok {
		logger.Debug("the ref is not a branch", "ref", ref)
		return 0, nil
	}
	branch2, ok := strings.CutPrefix(branch, "gh-readonly-queue/")
	if !ok {
		logger.Debug("the branch is not a gh-readonly-queue", "branch", branch)
		return 0, nil
	}
	// e.g. pr-24-a9d10f59f8c051673f45263c42aca8346614e716
	s, _, ok := strings.Cut(strings.TrimPrefix(path.Base(branch2), "pr-"), "-")
	if !ok {
		return 0, errors.New("gh-readonly-queue branch is not a valid format")
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse pull request number in gh-readonly-queue branch as number: %w", err)
	}
	return n, nil
}

func newMergeGroupEvent(logger *slog.Logger, ev *github.MergeGroupEvent) (*Event, error) {
	// e.g. refs/heads/gh-readonly-queue/main/pr-24-a9d10f59f8c051673f45263c42aca8346614e716
	prNumber, err := getPRNumberFromPushBranch(logger, ev.GetMergeGroup().GetHeadRef())
	if err != nil {
		return nil, fmt.Errorf("get a pull request number from the branch name: %w", err)
	}
	if prNumber == 0 {
		// Ignore webhook events not from gh-readonly-queue branches
		return nil, nil //nolint:nilnil
	}
	return &Event{
		Action:       ev.GetAction(),
		RepoFullName: ev.GetRepo().GetFullName(),
		RepoOwner:    ev.GetRepo().GetOwner().GetLogin(),
		RepoName:     ev.GetRepo().GetName(),
		PRNumber:     prNumber,
		RepoID:       ev.GetRepo().GetNodeID(),
		HeadSHA:      ev.GetMergeGroup().GetHeadSHA(),
	}, nil
}
