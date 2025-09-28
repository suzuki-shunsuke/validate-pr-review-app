package gcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// SecretsManager wraps the Google Cloud Secret Manager client for retrieving application secrets.
type SecretsManager struct {
	client *secretmanager.Client
}

// newSecretManager creates a new SecretsManager instance with a Google Cloud Secret Manager client.
// It initializes the client using the default Google Cloud credentials.
func newSecretManager(ctx context.Context) (*SecretsManager, error) {
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create secret manager client: %w", err)
	}
	return &SecretsManager{client: c}, nil
}

// Secret represents the structure of secrets stored in Google Cloud Secret Manager.
// It contains the GitHub App private key and webhook secret required for the application.
type Secret struct {
	GitHubAppPrivateKey string `json:"github_app_private_key"`
	WebhookSecret       string `json:"webhook_secret"`
}

// Validate checks if the Secret contains all required fields.
// It returns an error if any required field is missing or empty.
func (s *Secret) Validate() error {
	if s == nil {
		return errors.New("Secret is nil")
	}
	if s.GitHubAppPrivateKey == "" {
		return errors.New("GitHubAppPrivateKey is required")
	}
	if s.WebhookSecret == "" {
		return errors.New("WebhookSecret is required")
	}
	return nil
}

// Get retrieves a secret from Google Cloud Secret Manager and unmarshals it into a Secret struct.
// The secret value is expected to be a JSON string containing the required fields.
func (sm *SecretsManager) Get(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*Secret, error) {
	resp, err := sm.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("access secret version: %w", err)
	}
	data := resp.GetPayload().GetData()
	if data == nil {
		return nil, errors.New("secret payload data is nil")
	}
	var secret Secret
	if err := json.Unmarshal(data, &secret); err != nil {
		return nil, fmt.Errorf("unmarshal secret as JSON: %w", err)
	}
	return &secret, nil
}

// readSecret is a convenience function that creates a SecretsManager client
// and retrieves the secret with the given ID from Google Cloud Secret Manager.
func readSecret(ctx context.Context, secretID string) (*Secret, error) {
	sm, err := newSecretManager(ctx)
	if err != nil {
		return nil, err
	}
	defer sm.client.Close()
	secret, err := sm.Get(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretID,
	})
	if err != nil {
		return nil, err
	}
	return secret, nil
}
