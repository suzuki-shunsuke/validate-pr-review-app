package secret

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Secret struct {
	GitHubAppPrivateKey string `json:"github_app_private_key" yaml:"github_app_private_key"`
	WebhookSecret       string `json:"webhook_secret" yaml:"webhook_secret"`
}

// Validate validates the secret.
// useKMS is true if JWTs are signed by AWS KMS.
// Then the GitHub App's private key must not be set, otherwise it's required.
func (s *Secret) Validate(useKMS bool) error {
	if s == nil {
		return errors.New("Secret is nil")
	}
	if useKMS && s.GitHubAppPrivateKey != "" {
		return errors.New("GitHubAppPrivateKey must not be set when aws.kms_key_id is set")
	}
	if !useKMS && s.GitHubAppPrivateKey == "" {
		return errors.New("GitHubAppPrivateKey is required unless aws.kms_key_id is set")
	}
	if s.WebhookSecret == "" {
		return errors.New("WebhookSecret is required")
	}
	return nil
}

func Read(secret *Secret) error {
	if p := os.Getenv("SECRET_FILE"); p != "" {
		data, err := os.ReadFile(p) //nolint:gosec
		if err != nil {
			return fmt.Errorf("read secret file %q: %w", p, err)
		}
		if err := yaml.Unmarshal(data, secret); err != nil {
			return fmt.Errorf("unmarshal secret file: %w", err)
		}
	}
	if v := os.Getenv("GITHUB_APP_PRIVATE_KEY"); v != "" {
		secret.GitHubAppPrivateKey = v
	}
	if v := os.Getenv("WEBHOOK_SECRET"); v != "" {
		secret.WebhookSecret = v
	}
	return nil
}
