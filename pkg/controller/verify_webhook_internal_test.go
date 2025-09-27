package controller

import (
	//nolint:gosec
	"crypto/hmac"
	"crypto/sha1"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
)

// generateSignature creates a valid HMAC-SHA1 signature for testing
func generateSignature(payload string, secret []byte) string {
	h := hmac.New(sha1.New, secret)
	h.Write([]byte(payload))
	return fmt.Sprintf("sha1=%x", h.Sum(nil))
}

func TestHandler_validateRequest(t *testing.T) { //nolint:gocognit,cyclop
	t.Parallel()

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
	validSignature := generateSignature(validPayload, validSecret)

	tests := []struct {
		name          string
		controller    *Controller
		request       *Request
		wantErr       error
		wantPayload   bool
		expectedEvent *Event
	}{
		{
			name: "missing X-HUB-SIGNATURE header",
			controller: &Controller{
				input: &InputNew{
					Config: &config.Config{AppID: 12345},
				},
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
			},
			request: &Request{
				Body: validPayload,
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "12345",
						headerXHubSignature:                   validSignature,
						headerXGitHubEvent:                    eventPullRequestReview,
					},
				},
			},
			wantErr: errSignatureInvalid,
		},
		{
			name: "missing X-GITHUB-EVENT header",
			controller: &Controller{
				input: &InputNew{
					Config:        &config.Config{AppID: 12345},
					WebhookSecret: validSecret,
				},
			},
			request: &Request{
				Body: validPayload,
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "12345",
						headerXHubSignature:                   validSignature,
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
			},
			request: &Request{
				Body: validPayload,
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "12345",
						headerXHubSignature:                   validSignature,
						headerXGitHubEvent:                    "push",
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
			},
			request: &Request{
				Body: "invalid json{",
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "12345",
						headerXHubSignature:                   generateSignature("invalid json{", []byte("test-secret")),
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
			},
			request: &Request{
				Body: validPayload,
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "12345",
						headerXHubSignature:                   validSignature,
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
