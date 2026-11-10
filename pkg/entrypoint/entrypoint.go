package entrypoint

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/aws"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/gcloud"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/logging"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/secret"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/server"
)

func Run(ctx context.Context, logger *slog.Logger, logLevel *slog.LevelVar, getEnv func(string) string, version string) error {
	cfg := &config.Config{}
	if err := config.Read(cfg); err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	if err := logging.SetLevel(logLevel, cfg.LogLevel); err != nil {
		return fmt.Errorf("set log level: %w", err)
	}
	s, err := readSecret(ctx, cfg)
	if err != nil {
		return err
	}
	if err := s.Validate(cfg.UseAWSKMS()); err != nil {
		return fmt.Errorf("validate secret: %w", err)
	}
	signer, err := newSigner(ctx, cfg, s)
	if err != nil {
		return err
	}
	ctrl, err := controller.New(&controller.InputNew{
		Config:          cfg,
		Version:         version,
		WebhookSecret:   []byte(s.WebhookSecret),
		GitHubAppSigner: signer,
		Logger:          logger,
	})
	if err != nil {
		return fmt.Errorf("create controller: %w", err)
	}

	if getEnv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// lambda
		handler, err := aws.NewHandler(logger, ctrl, cfg)
		if err != nil {
			return fmt.Errorf("create a new handler: %w", err)
		}
		handler.Start(ctx)
		return nil
	}

	// http server
	server, err := server.New(logger, ctrl, cfg)
	if err != nil {
		return fmt.Errorf("create a new server: %w", err)
	}
	server.Start(ctx)
	return nil
}

// newSigner creates a Signer of JWTs for the GitHub App.
// If aws.kms_key_id is set, JWTs are signed by AWS KMS.
// Otherwise, they are signed with the GitHub App's private key.
func newSigner(ctx context.Context, cfg *config.Config, s *secret.Secret) (github.Signer, error) {
	if cfg.UseAWSKMS() {
		signer, err := aws.NewKMSSigner(ctx, cfg.AWS.KMSKeyID)
		if err != nil {
			return nil, fmt.Errorf("create a signer with AWS KMS: %w", err)
		}
		return signer, nil
	}
	signer, err := github.NewRSASigner(s.GitHubAppPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("create a signer with the GitHub App's private key: %w", err)
	}
	return signer, nil
}

func readSecret(ctx context.Context, cfg *config.Config) (*secret.Secret, error) {
	if cfg.AWS != nil && cfg.AWS.SecretID != "" {
		secret, err := aws.ReadSecret(ctx, cfg.AWS.SecretID)
		if err != nil {
			return nil, fmt.Errorf("get secret from AWS Secrets Manager: %w", err)
		}
		return secret, nil
	}
	if cfg.GoogleCloud != nil && cfg.GoogleCloud.SecretName != "" {
		secret, err := gcloud.ReadSecret(ctx, cfg.GoogleCloud.SecretName)
		if err != nil {
			return nil, fmt.Errorf("get secret from Google Cloud SecretManager: %w", err)
		}
		return secret, nil
	}
	s := &secret.Secret{}
	if err := secret.Read(s); err != nil {
		return nil, fmt.Errorf("read secret: %w", err)
	}
	return s, nil
}
