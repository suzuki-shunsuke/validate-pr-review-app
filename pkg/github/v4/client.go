package v4

import (
	"net/http"

	"github.com/google/go-github/v93/github"
	"github.com/shurcooL/githubv4"
)

type Client struct {
	v4Client *githubv4.Client
}

type PullRequestReviewEvent = github.PullRequestReviewEvent

// New creates a GitHub GraphQL API client.
// httpClient must handle authentication, retries, and timeouts.
func New(httpClient *http.Client) *Client {
	return &Client{
		v4Client: githubv4.NewClient(httpClient),
	}
}
