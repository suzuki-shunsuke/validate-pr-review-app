//nolint:funlen
package controller

import (
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
)

func newMockValidateSignature(err error) func(_ string, _, _ []byte) error {
	return func(_ string, _, _ []byte) error {
		return err
	}
}

func TestHandler_validateRequest(t *testing.T) { //nolint:gocognit,cyclop
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
		wantErr       error
		wantErrMsg    string
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
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "12345",
						headerXGitHubEvent:                    eventPullRequestReview,
					},
				},
			},
			wantErr: errHeaderXHubSignatureIsRequired,
		},
		{
			name: "invalid signature",
			controller: &Controller{
				input: &InputNew{
					Config:        &config.Config{AppID: 12345},
					WebhookSecret: []byte("wrong-secret"),
				},
				validateSignature: newMockValidateSignature(errors.New("invalid signature")),
			},
			request: &Request{
				Body: validPayload,
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "12345",
						headerXHubSignature:                   dummySignature,
						headerXGitHubEvent:                    eventPullRequestReview,
					},
				},
			},
			wantErrMsg: "invalid signature",
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
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "12345",
						headerXHubSignature:                   dummySignature,
					},
				},
			},
			wantErr: errHeaderXHubEventIsRequired,
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
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "12345",
						headerXHubSignature:                   dummySignature,
						headerXGitHubEvent:                    "label",
					},
				},
			},
			wantErr: errInvalidEventType,
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
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "12345",
						headerXHubSignature:                   dummySignature,
						headerXGitHubEvent:                    eventPullRequestReview,
					},
				},
			},
			wantErr: nil, // This will be a dynamic error message about JSON parsing
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
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "12345",
						headerXHubSignature:                   dummySignature,
						headerXGitHubEvent:                    eventPullRequestReview,
					},
				},
			},
			wantPayload: true,
		},
		{
			name: "empty headers",
			request: &Request{
				Body: "{}",
				Params: &RequestParamsField{
					Headers: map[string]string{},
				},
			},
			wantErr: errHeaderXHubSignatureIsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			payload, err := tt.controller.verifyWebhook(logger, tt.request)

			// Check error expectations
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("validateRequest() expected error %v, got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("validateRequest() error = %v, want %v", err, tt.wantErr)
					return
				}
				return
			}
			if tt.wantErrMsg != "" {
				if err == nil {
					t.Errorf("validateRequest() expected error containing %q, got nil", tt.wantErrMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("validateRequest() error = %v, want error containing %q", err, tt.wantErrMsg)
					return
				}
				return
			}

			// For cases where we expect specific error messages but not specific error types
			if !tt.wantPayload && err == nil {
				t.Error("validateRequest() expected an error, got nil")
				return
			}

			// For valid requests
			if tt.wantPayload {
				if err != nil {
					t.Errorf("validateRequest() unexpected error = %v", err)
					return
				}
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
