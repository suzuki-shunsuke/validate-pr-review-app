package github

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v93/github"
	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/go-retryablehttp"
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
)

// requestTimeout is the timeout of each attempt of an HTTP request to GitHub API.
// Each retry has its own timeout.
// It prevents a stalled connection from blocking webhook processing forever.
// GitHub terminates API requests that take more than 10 seconds to process,
// so waiting much longer than that is pointless.
// It is also short enough that a retry can happen within the HTTP server's graceful shutdown (25 seconds).
// https://docs.github.com/en/rest/using-the-rest-api/troubleshooting-the-rest-api
// https://docs.github.com/en/graphql/overview/rate-limits-and-query-limits-for-the-graphql-api
// Note that it doesn't bound the wait between retries:
// go-retryablehttp waits as long as the Retry-After header says on 429 and 503 responses.
const requestTimeout = 15 * time.Second

// ParamNewApp is the parameter of New.
// The GitHub API clients authenticate as the installation of the GitHub App.
type ParamNewApp struct {
	AppID          int64
	KeyFile        string
	InstallationID int64
	Logger         *slog.Logger
}

var ValidateSignature = github.ValidateSignature //nolint:gochecknoglobals

// New creates a client of GitHub REST and GraphQL APIs.
// Both share one HTTP client, so they share the authentication, retry, and timeout settings.
func New(param *ParamNewApp) (*Client, error) {
	httpClient, err := newHTTPClient(param)
	if err != nil {
		return nil, fmt.Errorf("create an HTTP client for GitHub API: %w", err)
	}
	v3Client, err := v3.New(httpClient)
	if err != nil {
		return nil, fmt.Errorf("create GitHub v3 client: %w", err)
	}
	return &Client{
		v4Client: v4.New(httpClient),
		v3Client: v3Client,
	}, nil
}

// newHTTPClient creates an HTTP client shared by the GitHub REST and GraphQL API clients.
// It authenticates as the GitHub App installation, retries failed requests, and times out each attempt.
func newHTTPClient(param *ParamNewApp) (*http.Client, error) {
	itr, err := ghinstallation.New(http.DefaultTransport, param.AppID, param.InstallationID, []byte(param.KeyFile))
	if err != nil {
		return nil, fmt.Errorf("create a transport with private key: %w", err)
	}
	c := retryablehttp.NewClient()
	c.HTTPClient = &http.Client{Transport: itr, Timeout: requestTimeout}
	c.Logger = param.Logger
	return c.StandardClient(), nil
}
