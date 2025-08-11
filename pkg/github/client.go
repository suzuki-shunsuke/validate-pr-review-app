package github

import (
	"context"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Client struct {
	v4Client *githubv4.Client
}

func (c *Client) Init(ctx context.Context, token string) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
	c.v4Client = githubv4.NewClient(httpClient)
}
