//nolint:funlen
package controller

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
)

func newMockValidateSignature(err error) func(_ string, _, _ []byte) error {
	return func(_ string, _, _ []byte) error {
		return err
	}
}

func TestHandler_validateRequest(t *testing.T) {
	t.Parallel()
	const dummySignature = "sha256=abcdefghijklmnopqrstuvwxyz0123456789abcdef"

	// Create a test logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Valid GitHub webhook payload for testing
	validPayload := `{
		"action": "submitted",
		"pull_request": {
			"number": 123,
			"head": {
				"sha": "abc123"
			}
		},
		"review": {
			"state": "approved",
			"user": {
				"login": "reviewer"
			}
		}
	}`

	// Generate valid signature for testing
	validSecret := []byte("test-secret")

	tests := []struct {
		name          string
		controller    *Controller
		request       *Request
		wantPayload   bool
		expectedEvent *Event
	}{
		{
			name: "missing X-HUB-SIGNATURE header",
			controller: &Controller{
				input: &InputNew{
					Config: &config.Config{AppID: 12345},
				},
				validateSignature: newMockValidateSignature(nil),
			},
			request: &Request{
				Body: validPayload,
				Headers: map[string]string{
					headerXGitHubHookInstallationTargetID: "12345",
					headerXHubSignature:                   dummySignature,
					headerXGitHubEvent:                    eventPullRequestReview,
				},
			},
		},
		{
			name: "invalid X-HUB-SIGNATURE header",
			controller: &Controller{
				input: &InputNew{
					Config: &config.Config{AppID: 12345},
				},
				validateSignature: newMockValidateSignature(errors.New("invalid signature")),
			},
			request: &Request{
				Body: validPayload,
				Headers: map[string]string{
					headerXGitHubHookInstallationTargetID: "12345",
					headerXHubSignature:                   dummySignature,
					headerXGitHubEvent:                    eventPullRequestReview,
				},
			},
		},
		{
			name: "missing X-GITHUB-EVENT header",
			controller: &Controller{
				input: &InputNew{
					Config:        &config.Config{AppID: 12345},
					WebhookSecret: validSecret,
				},
				validateSignature: newMockValidateSignature(nil),
			},
			request: &Request{
				Body: validPayload,
				Headers: map[string]string{
					headerXGitHubHookInstallationTargetID: "12345",
					headerXHubSignature:                   dummySignature,
				},
			},
		},
		{
			name: "unsupported event type",
			controller: &Controller{
				input: &InputNew{
					Config:        &config.Config{AppID: 12345},
					WebhookSecret: validSecret,
				},
				validateSignature: newMockValidateSignature(nil),
			},
			request: &Request{
				Body: validPayload,
				Headers: map[string]string{
					headerXGitHubHookInstallationTargetID: "12345",
					headerXHubSignature:                   dummySignature,
					headerXGitHubEvent:                    "label",
				},
			},
		},
		{
			name: "invalid JSON payload",
			controller: &Controller{
				input: &InputNew{
					Config:        &config.Config{AppID: 12345},
					WebhookSecret: []byte("test-secret"),
				},
				validateSignature: newMockValidateSignature(nil),
			},
			request: &Request{
				Body: "invalid json{",
				Headers: map[string]string{
					headerXGitHubHookInstallationTargetID: "12345",
					headerXHubSignature:                   dummySignature,
					headerXGitHubEvent:                    eventPullRequestReview,
				},
			},
		},
		{
			name: "valid request",
			controller: &Controller{
				input: &InputNew{
					Config:        &config.Config{AppID: 12345},
					WebhookSecret: validSecret,
				},
				validateSignature: newMockValidateSignature(nil),
			},
			request: &Request{
				Body: validPayload,
				Headers: map[string]string{
					headerXGitHubHookInstallationTargetID: "12345",
					headerXHubSignature:                   dummySignature,
					headerXGitHubEvent:                    eventPullRequestReview,
				},
			},
			wantPayload: true,
		},
		{
			name: "empty headers",
			request: &Request{
				Body:    "{}",
				Headers: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			payload := tt.controller.verifyWebhook(logger, tt.request)
			if tt.wantPayload {
				if payload == nil {
					t.Error("validateRequest() returned nil payload")
					return
				}
				// Verify it's a valid Event
				if payload.Action == "" {
					t.Error("verifyWebhook() returned payload without Action field")
				}
			}
		})
	}
}

func Test_getPRNumberFromBranch(t *testing.T) {
	t.Parallel()

	logger := slog.Default()

	tests := []struct {
		name           string
		branch         string
		expectedNumber int
		wantErr        bool
	}{
		{
			name:           "valid gh-readonly-queue branch",
			branch:         "gh-readonly-queue/main/pr-24-a9d10f59f8c051673f45263c42aca8346614e716",
			expectedNumber: 24,
			wantErr:        false,
		},
		{
			name:           "not a gh-readonly-queue branch",
			branch:         "main",
			expectedNumber: 0,
			wantErr:        false,
		},
		{
			name:           "gh-readonly-queue but invalid format - missing pr prefix",
			branch:         "gh-readonly-queue/main/24-a9d10f59f8c051673f45263c42aca8346614e716",
			expectedNumber: 0,
			wantErr:        true,
		},
		{
			name:           "gh-readonly-queue but invalid format - no dash after pr number",
			branch:         "gh-readonly-queue/main/pr-24a9d10f59f8c051673f45263c42aca8346614e716",
			expectedNumber: 0,
			wantErr:        true,
		},
		{
			name:           "gh-readonly-queue but invalid format - non-numeric PR number",
			branch:         "gh-readonly-queue/main/pr-abc-a9d10f59f8c051673f45263c42aca8346614e716",
			expectedNumber: 0,
			wantErr:        true,
		},
		{
			name:           "gh-readonly-queue but empty PR number",
			branch:         "gh-readonly-queue/main/pr--a9d10f59f8c051673f45263c42aca8346614e716",
			expectedNumber: 0,
			wantErr:        true,
		},
		{
			name:           "gh-readonly-queue with refs/heads prefix",
			branch:         "refs/heads/gh-readonly-queue/main/pr-42-fedcba9876543210",
			expectedNumber: 0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			number, err := getPRNumberFromBranch(logger, tt.branch)

			if (err != nil) != tt.wantErr {
				t.Errorf("getPRNumberFromBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if number != tt.expectedNumber {
				t.Errorf("getPRNumberFromBranch() = %v, want %v", number, tt.expectedNumber)
			}
		})
	}
}
