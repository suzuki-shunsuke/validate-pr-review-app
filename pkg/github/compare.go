package github

import (
	"context"
	"fmt"
)

// CompareCommits returns the list of changed file paths between two commits.
func (c *Client) CompareCommits(ctx context.Context, owner, repo, base, head string) ([]string, error) {
	files, err := c.v3Client.CompareCommits(ctx, owner, repo, base, head)
	if err != nil {
		return nil, fmt.Errorf("compare commits: %w", err)
	}
	return files, nil
}
