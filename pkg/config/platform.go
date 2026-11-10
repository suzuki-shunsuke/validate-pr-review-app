package config

import (
	"errors"
	"fmt"
)

type AWS struct {
	// SecretID is the secret ID in AWS Secrets Manager.
	// If it's empty, secrets are read from environment variables or a file.
	SecretID             string `json:"secret_id,omitempty" yaml:"secret_id"`
	UseLambdaFunctionURL bool   `json:"use_lambda_function_url,omitempty" yaml:"use_lambda_function_url"`
	// KMSKeyID is the ID of the AWS KMS key where the GitHub App's private key is imported.
	// It's a key ID, key ARN, alias name, or alias ARN.
	// If it's set, JWTs are signed by AWS KMS and the GitHub App's private key must not be set.
	KMSKeyID string `json:"kms_key_id,omitempty" yaml:"kms_key_id"`
}

type GoogleCloud struct {
	SecretName string `json:"secret_name" yaml:"secret_name"`
}

func (g *GoogleCloud) Validate() error {
	if g.SecretName == "" {
		return errors.New("secret_name is required")
	}
	return nil
}

// UseAWSKMS returns true if JWTs are signed by AWS KMS.
func (c *Config) UseAWSKMS() bool {
	return c.AWS != nil && c.AWS.KMSKeyID != ""
}

func (c *Config) validatePlatform() error {
	if c.GoogleCloud != nil {
		if err := c.GoogleCloud.Validate(); err != nil {
			return fmt.Errorf("validate google_cloud config: %w", err)
		}
	}
	if c.AWS != nil && c.GoogleCloud != nil {
		return errors.New("only one of aws or google_cloud configuration can be set")
	}
	return nil
}
