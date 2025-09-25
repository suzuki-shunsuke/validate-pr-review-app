package v4

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

func (c *Client) CreateCheckRun(ctx context.Context, input githubv4.CreateCheckRunInput) error {
	var m struct {
		CreateCheckRun struct {
			CheckRun struct {
				ID githubv4.String
			}
		} `graphql:"createCheckRun(input:$input)"`
	}

	if err := c.v4Client.Mutate(ctx, &m, input, nil); err != nil {
		return fmt.Errorf("create a check run: %w", err)
	}
	return nil
}
