package aws

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"text/template"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/validation"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"gopkg.in/yaml.v3"
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
	// ListReviews(ctx context.Context, owner, name string, number int, cursor string) ([]*github.Review, error)
	// ListCommits(ctx context.Context, owner, name string, number int, cursor string) ([]*github.PullRequestCommit, error)
	CreateCheckRun(ctx context.Context, input githubv4.CreateCheckRunInput) error
}

type Validator interface {
	Run(logger *slog.Logger, input *validation.Input) *config.Result
}

func NewHandler(ctx context.Context, logger *slog.Logger) (*Handler, error) {
	// read config from the environment variable
	// parse config as YAML
	cfg := &config.Config{}
	if err := readConfig(cfg); err != nil {
		return nil, err
	}
	// Read AWS config
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
	// Create GitHub client
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
		validator: validation.New(&validation.InputNew{
			TrustedApps:           cfg.UniqueTrustedApps,
			UntrustedMachineUsers: cfg.UniqueUntrustedMachineUsers,
			TrustedMachineUsers:   cfg.UniqueTrustedMachineUsers,
		}),
		webhookSecret: []byte(secret.WebhookSecret),
	}, nil
}

func (h *Handler) Start(ctx context.Context) {
	var handler any
	if h.config.AWS.UseLambdaFunctionURL {
		handler = h.handleFunctionURL
	} else {
		handler = h.do
	}
	lambda.StartWithOptions(handler, lambda.WithContext(ctx))
}

func (h *Handler) do(ctx context.Context, req *Request) {
	logger := h.logger
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		logger.Warn("lambda context is not found")
	} else {
		logger = logger.With("aws_request_id", lc.AwsRequestID)
	}
	if err := h.handle(ctx, logger, req); err != nil {
		slogerr.WithError(logger, err).Error("handle request")
	}
}

func (h *Handler) validate(ctx context.Context, ev *github.PullRequestReviewEvent) *config.Result {
	repo := ev.GetRepo()
	owner := repo.GetOwner().GetLogin()
	pr, err := h.gh.GetPR(ctx, owner, repo.GetName(), ev.GetPullRequest().GetNumber())
	if err != nil {
		return &config.Result{Error: fmt.Errorf("get a pull request: %w", err).Error()}
	}
	h.logger.Info("Fetched a pull request", "pull_request", pr)
	return h.validator.Run(h.logger, &validation.Input{
		PR: pr,
	})
}

func summarize(result *config.Result, templates map[string]*template.Template) (string, error) {
	var key string
	if result.Error != "" {
		key = "error"
	} else {
		key = string(result.State)
	}
	tpl, ok := templates[key]
	if !ok {
		return "", errors.New("summary template is not found")
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, result); err != nil {
		return "", fmt.Errorf("execute summary template: %w", err)
	}
	return buf.String(), nil
}

func readConfig(cfg *config.Config) error {
	cfgstr := os.Getenv("CONFIG")
	if cfgstr == "" {
		return errors.New("CONFIG environment variable is required")
	}
	if err := yaml.Unmarshal([]byte(cfgstr), cfg); err != nil {
		return fmt.Errorf("failed to parse CONFIG environment variable: %w", err)
	}
	if err := cfg.Init(); err != nil {
		return fmt.Errorf("initialize config: %w", err)
	}
	return nil
}
