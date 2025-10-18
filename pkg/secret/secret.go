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

func Read(secret *Secret) error {
	if p := os.Getenv("SECRET_FILE"); p != "" {
		data, err := os.ReadFile(p)
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
