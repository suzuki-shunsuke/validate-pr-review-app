package v3

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/v93/github"
)

type Client struct {
	client *github.Client
}

// New creates a GitHub REST API client.
// httpClient must handle authentication, retries, and timeouts.
func New(httpClient *http.Client) (*Client, error) {
	gh, err := github.NewClient(github.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create a GitHub client: %w", err)
	}
	return &Client{
		client: gh,
	}, nil
}
