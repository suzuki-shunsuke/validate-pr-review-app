package aws

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/validation"
)

type Handler struct {
	logger        *slog.Logger
	webhookSecret []byte
	validator     Validator
	config        *config.Config
	gh            GitHub
}

type GitHub interface {
	GetPR(ctx context.Context, owner, name string, number int) (*github.PullRequest, error)
	ListReviews(ctx context.Context, owner, name string, number int, cursor string) ([]*github.Review, error)
	ListCommits(ctx context.Context, owner, name string, number int, cursor string) ([]*github.PullRequestCommit, error)
}

type Validator interface {
	Run(ctx context.Context, logger *slog.Logger, input *validation.Input) error
}

func NewHandler(ctx context.Context, logger *slog.Logger) (*Handler, error) {
	// read config from the environment variable
	// parse config as YAML
	cfg := &config.Config{}
	if err := readConfig(cfg); err != nil {
		return nil, err
	}
	config, err := NewConfig(ctx)
	if err != nil {
		return nil, err
	}
	// read secrets from AWS SecretsManager
	sm := NewSecretsManager(config)
	secret, err := sm.Get(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(cfg.AWS.SecretID),
	})
	if err != nil {
		return nil, fmt.Errorf("get secret from AWS Secrets Manager: %w", err)
	}
	gh, err := github.New(&github.ParamNewApp{
		AppID:          cfg.AppID,
		InstallationID: cfg.InstallationID,
		KeyFile:        secret.GitHubAppPrivateKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create a GitHub client: %w", err)
	}
	return &Handler{
		logger: logger,
		config: cfg,
		gh:     gh,
	}, nil
}

func (h *Handler) Start(ctx context.Context) {
	lambda.StartWithOptions(h.do, lambda.WithContext(ctx))
}

func (h *Handler) do(ctx context.Context, req *Request) error {
	h.logger.Info("Starting a request", "request", req)
	defer h.logger.Info("Ending a request", "request", req)
	// TODO parse webhook payload
	ev, err := h.validate(h.logger, req)
	if err != nil {
		h.logger.Warn("Failed to validate request", "error", err)
		return err
	}
	// TODO process the event
	if err := h.validator.Run(ctx, h.logger, &validation.Input{
		RepoOwner:             ev.GetRepo().GetOwner().GetLogin(),
		RepoName:              ev.GetRepo().GetName(),
		PR:                    ev.GetPullRequest().GetNumber(),
		TrustedApps:           h.config.TrustedApps,
		UntrustedMachineUsers: h.config.UntrustedMachineUsers,
	}); err != nil {
		h.logger.Error("Failed to run validation", "error", err)
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

func readConfig(cfg *config.Config) error {
	cfgstr := os.Getenv("CONFIG")
	if cfgstr == "" {
		return errors.New("CONFIG environment variable is required")
	}
	if err := yaml.Unmarshal([]byte(cfgstr), cfg); err != nil {
		return fmt.Errorf("failed to parse CONFIG environment variable: %w", err)
	}
	return nil
}
