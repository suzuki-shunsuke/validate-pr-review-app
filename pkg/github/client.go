package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v81/github"
	"github.com/shurcooL/githubv4"
	v4 "github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github/v4"
)

type Client struct {
	v4Client V4Client
}

type V4Client interface {
	GetPR(ctx context.Context, owner, name string, number int) (*v4.PullRequest, error)
	CreateCheckRun(ctx context.Context, input githubv4.CreateCheckRunInput) error
}

type (
	PullRequestReviewEvent = github.PullRequestReviewEvent
	CheckSuiteEvent        = github.CheckSuiteEvent
	ParamNewApp            = v4.ParamNewApp
)

var ValidateSignature = github.ValidateSignature //nolint:gochecknoglobals

func New(param *v4.ParamNewApp) (*Client, error) {
	v4Client, err := v4.New(param)
	if err != nil {
		return nil, fmt.Errorf("create GitHub v4 client: %w", err)
	}
	return &Client{
		v4Client: v4Client,
	}, nil
}
