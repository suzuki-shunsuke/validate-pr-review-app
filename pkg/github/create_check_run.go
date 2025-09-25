package github

import (
	"context"

	"github.com/shurcooL/githubv4"
)

func (c *Client) CreateCheckRun(ctx context.Context, input githubv4.CreateCheckRunInput) error {
	if err := c.v4Client.CreateCheckRun(ctx, input); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}
