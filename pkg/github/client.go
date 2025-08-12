package github

import (
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/shurcooL/githubv4"
)

type Client struct {
	v4Client *githubv4.Client
}

func New(param *ParamNewApp) (*Client, error) {
	itr, err := ghinstallation.New(http.DefaultTransport, param.AppID, param.InstallationID, []byte(param.KeyFile))
	if err != nil {
		return nil, fmt.Errorf("create a transport with private key: %w", err)
	}
	return &Client{
		v4Client: githubv4.NewClient(&http.Client{Transport: itr}),
	}, nil
}

type ParamNewApp struct {
	AppID          int64
	KeyFile        string
	InstallationID int64
}
