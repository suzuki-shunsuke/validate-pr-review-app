package v4

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v93/github"
	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/go-retryablehttp"
)

// requestTimeout is the timeout of each attempt of an HTTP request to GitHub API. Each retry has its own timeout.
const requestTimeout = 30 * time.Second

type Client struct {
	v4Client *githubv4.Client
}

type PullRequestReviewEvent = github.PullRequestReviewEvent

func New(param *ParamNewApp) (*Client, error) {
	itr, err := ghinstallation.New(http.DefaultTransport, param.AppID, param.InstallationID, []byte(param.KeyFile))
	if err != nil {
		return nil, fmt.Errorf("create a transport with private key: %w", err)
	}
	c := retryablehttp.NewClient()
	// Set a timeout so that a stalled request doesn't block processing forever.
	// The HTTP server waits for in-flight webhook events on shutdown without a deadline.
	c.HTTPClient = &http.Client{Transport: itr, Timeout: requestTimeout}
	c.Logger = param.Logger
	return &Client{
		v4Client: githubv4.NewClient(c.StandardClient()),
	}, nil
}

type ParamNewApp struct {
	AppID          int64
	KeyFile        string
	InstallationID int64
	Logger         *slog.Logger
}
