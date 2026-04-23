package v3

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v85/github"
	"github.com/suzuki-shunsuke/go-retryablehttp"
)

type Client struct {
	client *github.Client
}

type ParamNewApp struct {
	AppID          int64
	KeyFile        string
	InstallationID int64
	Logger         *slog.Logger
}

func New(param *ParamNewApp) (*Client, error) {
	itr, err := ghinstallation.New(http.DefaultTransport, param.AppID, param.InstallationID, []byte(param.KeyFile))
	if err != nil {
		return nil, fmt.Errorf("create a transport with private key: %w", err)
	}
	c := retryablehttp.NewClient()
	c.HTTPClient = &http.Client{Transport: itr}
	c.Logger = param.Logger
	return &Client{
		client: github.NewClient(c.StandardClient()),
	}, nil
}
