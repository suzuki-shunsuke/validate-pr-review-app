package v4

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v83/github"
	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/go-retryablehttp"
)

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
	c.HTTPClient = &http.Client{Transport: itr}
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
