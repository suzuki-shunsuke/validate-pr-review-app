package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v84/github"
	"github.com/shurcooL/githubv4"
	v3 "github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github/v3"
	v4 "github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github/v4"
)

type Client struct {
	v4Client V4Client
	v3Client V3Client
}

type V4Client interface {
	GetPR(ctx context.Context, owner, name string, number int) (*v4.PullRequest, error)
	CreateCheckRun(ctx context.Context, input githubv4.CreateCheckRunInput) error
}

type V3Client interface {
	CompareCommits(ctx context.Context, owner, repo, base, head string) ([]string, error)
	IsAncestor(ctx context.Context, owner, repo, ancestor, descendant string) (bool, error)
}

type (
	PullRequestReviewEvent = github.PullRequestReviewEvent
	PullRequestEvent       = github.PullRequestEvent
	CheckSuiteEvent        = github.CheckSuiteEvent
	ParamNewApp            = v4.ParamNewApp
)

var ValidateSignature = github.ValidateSignature //nolint:gochecknoglobals

func New(param *v4.ParamNewApp) (*Client, error) {
	v4Client, err := v4.New(param)
	if err != nil {
		return nil, fmt.Errorf("create GitHub v4 client: %w", err)
	}
	v3Client, err := v3.New(&v3.ParamNewApp{
		AppID:          param.AppID,
		InstallationID: param.InstallationID,
		KeyFile:        param.KeyFile,
		Logger:         param.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("create GitHub v3 client: %w", err)
	}
	return &Client{
		v4Client: v4Client,
		v3Client: v3Client,
	}, nil
}
