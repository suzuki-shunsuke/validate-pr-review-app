package gcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type SecretsManager struct {
	client *secretmanager.Client
}

func newSecretManager(ctx context.Context) (*SecretsManager, error) {
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create secret manager client: %w", err)
	}
	return &SecretsManager{client: c}, nil
}

type Secret struct {
	GitHubAppPrivateKey string `json:"github_app_private_key"`
	WebhookSecret       string `json:"webhook_secret"`
}

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
