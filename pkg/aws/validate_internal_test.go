//nolint:funlen
package aws

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-github/v74/github"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/config"
)

// generateSignature creates a valid HMAC-SHA1 signature for testing
func generateSignature(payload string, secret []byte) string {
	h := hmac.New(sha1.New, secret)
	h.Write([]byte(payload))
	return fmt.Sprintf("sha1=%x", h.Sum(nil))
}

func TestHandler_validateRequest(t *testing.T) { //nolint:gocognit,cyclop,maintidx
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
		handler       *Handler
		request       *Request
		wantErr       error
		wantPayload   bool
		expectedEvent *github.PullRequestReviewEvent
	}{
		{
			name: "missing X-GITHUB-HOOK-INSTALLATION-TARGET-ID header",
			handler: &Handler{
				config: &config.Config{AppID: 12345},
			},
			request: &Request{
				Body: validPayload,
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXHubSignature: validSignature,
						headerXGitHubEvent:  eventPullRequestReview,
					},
				},
			},
			wantErr: errHeaderXGitHubHookInstallationTargetIDIsRequired,
		},
		{
			name: "invalid X-GITHUB-HOOK-INSTALLATION-TARGET-ID header (non-integer)",
			handler: &Handler{
				config: &config.Config{AppID: 12345},
			},
			request: &Request{
				Body: validPayload,
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "invalid",
						headerXHubSignature:                   validSignature,
						headerXGitHubEvent:                    eventPullRequestReview,
					},
				},
			},
			wantErr: errHeaderXGitHubHookInstallationTargetIDMustBeInt64,
		},
		{
			name: "mismatched app ID",
			handler: &Handler{
				config:        &config.Config{AppID: 12345},
				webhookSecret: validSecret,
			},
			request: &Request{
				Body: validPayload,
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "99999",
						headerXHubSignature:                   validSignature,
						headerXGitHubEvent:                    eventPullRequestReview,
					},
				},
			},
			wantErr: nil, // This will be a dynamic error message
		},
		{
			name: "missing X-HUB-SIGNATURE header",
			handler: &Handler{
				config: &config.Config{AppID: 12345},
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
			handler: &Handler{
				config:        &config.Config{AppID: 12345},
				webhookSecret: []byte("wrong-secret"),
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
			handler: &Handler{
				config:        &config.Config{AppID: 12345},
				webhookSecret: validSecret,
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
			handler: &Handler{
				config:        &config.Config{AppID: 12345},
				webhookSecret: validSecret,
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
			handler: &Handler{
				config:        &config.Config{AppID: 12345},
				webhookSecret: []byte("test-secret"),
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
			handler: &Handler{
				config:        &config.Config{AppID: 12345},
				webhookSecret: validSecret,
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
			name: "app ID mismatch error message",
			handler: &Handler{
				config:        &config.Config{AppID: 12345},
				webhookSecret: []byte("test-secret"),
			},
			request: &Request{
				Body: "{}",
				Params: &RequestParamsField{
					Headers: map[string]string{
						headerXGitHubHookInstallationTargetID: "99999",
						headerXHubSignature:                   generateSignature("{}", []byte("test-secret")),
						headerXGitHubEvent:                    eventPullRequestReview,
					},
				},
			},
			wantErr: errInvalidAppID,
		},
		{
			name: "empty headers",
			request: &Request{
				Body: "{}",
				Params: &RequestParamsField{
					Headers: map[string]string{},
				},
			},
			wantErr: errHeaderXGitHubHookInstallationTargetIDIsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			payload, err := tt.handler.validateRequest(logger, tt.request)

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
				// Verify it's a valid PullRequestReviewEvent
				if payload.Action == nil {
					t.Error("validateRequest() returned payload without Action field")
				}
			}
		})
	}
}
