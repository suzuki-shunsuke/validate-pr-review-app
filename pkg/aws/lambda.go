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
	ListReviews(ctx context.Context, owner, name string, number int, cursor string) ([]*github.Review, error)
	ListCommits(ctx context.Context, owner, name string, number int, cursor string) ([]*github.PullRequestCommit, error)
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
		logger:    logger,
		config:    cfg,
		gh:        gh,
		validator: validation.New(),
	}, nil
}

func (h *Handler) Start(ctx context.Context) {
	lambda.StartWithOptions(h.do, lambda.WithContext(ctx))
}

func (h *Handler) do(ctx context.Context, req *Request) {
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		h.logger.Warn("lambda context is not found")
	} else {
		h.logger = h.logger.With("aws_request_id", lc.AwsRequestID)
	}

	h.logger.Info("Starting a request", "request", req)
	defer h.logger.Info("Ending a request", "request", req)

	// Validate the request
	ev, err := h.validateRequest(h.logger, req)
	if err != nil {
		h.logger.Warn("Failed to validate request", "error", err)
		return
	}

	checkName := githubv4.String(h.config.CheckName)

	// Get repository ID for GraphQL mutation
	repoID := githubv4.String(ev.GetRepo().GetNodeID())
	headSha := githubv4.GitObjectID(ev.GetPullRequest().GetHead().GetSHA())

	// Create initial check run with IN_PROGRESS status
	// inProgressStatus := githubv4.RequestableCheckStatusStateInProgress
	// checkRunInput := githubv4.CreateCheckRunInput{
	// 	RepositoryID: repoID,
	// 	HeadSha:      headSha,
	// 	Name:         checkName,
	// 	Status:       &inProgressStatus,
	// 	Output: &githubv4.CheckRunOutput{
	// 		Title:   githubv4.String("Validating PR review requirements"),
	// 		Summary: githubv4.String("Checking if the PR meets review requirements..."),
	// 	},
	// }

	// if err := h.gh.CreateCheckRun(ctx, checkRunInput); err != nil {
	// 	h.logger.Error("Failed to create initial check run", "error", err)
	// 	// Continue with validation even if check run creation fails
	// }

	// Run validation
	var conclusion githubv4.CheckConclusionState
	var title, summary githubv4.String
	result := h.validate(ctx, ev)
	if result.Error != "" {
		conclusion = githubv4.CheckConclusionStateFailure
		title = githubv4.String("PR review requirements not met")
	} else {
		conclusion = githubv4.CheckConclusionStateSuccess
		title = githubv4.String("PR review requirements met")
	}
	s, err := summarize(result, h.config.BuiltTemplates)
	if err != nil {
		slogerr.WithError(h.logger, err).Error("summarize the result")
	}
	summary = githubv4.String(s)

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
}

func (h *Handler) validate(ctx context.Context, ev *github.PullRequestReviewEvent) *config.Result {
	repo := ev.GetRepo()
	owner := repo.GetOwner().GetLogin()
	pr, err := h.gh.GetPR(ctx, owner, repo.GetName(), ev.GetPullRequest().GetNumber())
	if err != nil {
		return &config.Result{Error: fmt.Errorf("get a pull request: %w", err).Error()}
	}
	return h.validator.Run(h.logger, &validation.Input{
		PR:     pr,
		Config: h.config,
	})
}

func summarize(result *config.Result, templates map[string]*template.Template) (string, error) {
	var buf bytes.Buffer
	tpl, ok := templates[string(result.State)]
	if !ok {
		return "", errors.New("summary template is not found")
	}
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
