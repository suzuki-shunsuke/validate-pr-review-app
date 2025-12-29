package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/secret"
)

type SecretsManager struct {
	client *secretsmanager.Client
}

func NewConfig(ctx context.Context) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx) //nolint:wrapcheck
}

func NewSecretsManager(config aws.Config) *SecretsManager {
	client := secretsmanager.NewFromConfig(config)
	return &SecretsManager{client: client}
}

func (sm *SecretsManager) Get(ctx context.Context, input *secretsmanager.GetSecretValueInput) (*secret.Secret, error) {
	// config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Create Secrets Manager client
	// svc := secretsmanager.NewFromConfig(config)
	result, err := sm.client.GetSecretValue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("get secret value from AWS Secrets Manager: %w", err)
	}
	secret := &secret.Secret{}
	if err := json.Unmarshal([]byte(*result.SecretString), secret); err != nil {
		return nil, fmt.Errorf("unmarshal the secret as JSON: %w", err)
	}
	return secret, nil
}

func ReadSecret(ctx context.Context, secretID string) (*secret.Secret, error) {
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
