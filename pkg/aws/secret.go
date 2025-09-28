package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// SecretsManager wraps the AWS Secrets Manager client for retrieving application secrets.
type SecretsManager struct {
	client *secretsmanager.Client
}

// NewConfig creates a new AWS configuration by loading the default configuration.
// It uses environment variables and AWS profiles for authentication.
func NewConfig(ctx context.Context) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx) //nolint:wrapcheck
}

// NewSecretsManager creates a new SecretsManager instance with the provided AWS configuration.
func NewSecretsManager(config aws.Config) *SecretsManager {
	client := secretsmanager.NewFromConfig(config)
	return &SecretsManager{client: client}
}

// Secret represents the structure of secrets stored in AWS Secrets Manager.
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

// Get retrieves a secret from AWS Secrets Manager and unmarshals it into a Secret struct.
// The secret value is expected to be a JSON string containing the required fields.
func (sm *SecretsManager) Get(ctx context.Context, input *secretsmanager.GetSecretValueInput) (*Secret, error) {
	result, err := sm.client.GetSecretValue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("get secret value from AWS Secrets Manager: %w", err)
	}
	secret := &Secret{}
	if err := json.Unmarshal([]byte(*result.SecretString), secret); err != nil {
		return nil, fmt.Errorf("unmarshal the secret as JSON: %w", err)
	}
	return secret, nil
}

// readSecret is a convenience function that creates an AWS config, initializes a SecretsManager,
// and retrieves the secret with the given ID.
func readSecret(ctx context.Context, secretID string) (*Secret, error) {
	// Read AWS config
	config, err := NewConfig(ctx)
	if err != nil {
		return nil, err
	}
	// read secrets from AWS SecretsManager
	sm := NewSecretsManager(config)
	secret, err := sm.Get(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})
	if err != nil {
		return nil, fmt.Errorf("get secret from AWS Secrets Manager: %w", err)
	}
	return secret, nil
}
