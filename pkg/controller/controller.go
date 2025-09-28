package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/validation"
)

type Controller struct {
	input             *InputNew
	gh                GitHub
	validator         Validator
	validateSignature func(signature string, payload, secretToken []byte) error
}

func New(input *InputNew) (*Controller, error) {
	// Create GitHub client
	gh, err := github.New(&github.ParamNewApp{
		AppID:          input.Config.AppID,
		InstallationID: input.Config.InstallationID,
		KeyFile:        input.GitHubAppPrivateKey,
		Logger:         input.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("create GitHub client: %w", err)
	}
	return &Controller{
		input:             input,
		gh:                gh,
		validator:         validation.New(&validation.InputNew{}),
		validateSignature: github.ValidateSignature,
	}, nil
}

type InputNew struct {
	Config              *config.Config
	Version             string
	WebhookSecret       []byte
	GitHubAppPrivateKey string
	Logger              *slog.Logger
}

type Validator interface {
	Run(logger *slog.Logger, input *validation.Input) *validation.Result
}

type GitHub interface {
	GetPR(ctx context.Context, owner, name string, number int) (*github.PullRequest, error)
	CreateCheckRun(ctx context.Context, input githubv4.CreateCheckRunInput) error
}

type Request struct {
	// Generate template > Method request passthrough
	Body      string            `json:"body"`
	Headers   map[string]string `json:"header"`
	RequestID string            `json:"requestid"`
}
