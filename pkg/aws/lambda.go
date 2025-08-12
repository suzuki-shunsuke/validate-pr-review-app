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
	"github.com/shurcooL/githubv4"
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
	CreateCheckRun(ctx context.Context, input githubv4.CreateCheckRunInput) error
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

func (h *Handler) do(ctx context.Context, req *Request) error { //nolint:funlen
	h.logger.Info("Starting a request", "request", req)
	defer h.logger.Info("Ending a request", "request", req)
	ev, err := h.validate(h.logger, req)
	if err != nil {
		h.logger.Warn("Failed to validate request", "error", err)
		return err
	}

	// Create initial check run
	checkName := githubv4.String("Enforce PR Review")

	// Get repository ID for GraphQL mutation
	repoID := githubv4.String(ev.GetRepo().GetNodeID())
	headSha := githubv4.GitObjectID(ev.GetPullRequest().GetHead().GetSHA())

	// Create initial check run with IN_PROGRESS status
	inProgressStatus := githubv4.RequestableCheckStatusStateInProgress
	checkRunInput := githubv4.CreateCheckRunInput{
		RepositoryID: repoID,
		HeadSha:      headSha,
		Name:         checkName,
		Status:       &inProgressStatus,
		Output: &githubv4.CheckRunOutput{
			Title:   githubv4.String("Validating PR review requirements"),
			Summary: githubv4.String("Checking if the PR meets review requirements..."),
		},
	}

	if err := h.gh.CreateCheckRun(ctx, checkRunInput); err != nil {
		h.logger.Error("Failed to create initial check run", "error", err)
		// Continue with validation even if check run creation fails
	}

	// Run validation
	validationErr := h.run(ctx, ev)

	// Update check run based on validation result
	var conclusion githubv4.CheckConclusionState
	var title, summary githubv4.String

	if validationErr != nil {
		conclusion = githubv4.CheckConclusionStateFailure
		title = githubv4.String("PR review requirements not met")
		summary = githubv4.String(fmt.Sprintf("Validation failed: %v", validationErr))
	} else {
		conclusion = githubv4.CheckConclusionStateSuccess
		title = githubv4.String("PR review requirements met")
		summary = githubv4.String("All PR review requirements have been satisfied.")
	}

	// Create final check run with conclusion
	completedStatus := githubv4.RequestableCheckStatusStateCompleted
	finalCheckRunInput := githubv4.CreateCheckRunInput{
		RepositoryID: repoID,
		HeadSha:      headSha,
		Name:         checkName,
		Status:       &completedStatus,
		Conclusion:   &conclusion,
		Output: &githubv4.CheckRunOutput{
			Title:   title,
			Summary: summary,
		},
	}

	if err := h.gh.CreateCheckRun(ctx, finalCheckRunInput); err != nil {
		h.logger.Error("Failed to create final check run", "error", err)
	}

	return validationErr
}

func (h *Handler) run(ctx context.Context, ev *github.PullRequestReviewEvent) error {
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
