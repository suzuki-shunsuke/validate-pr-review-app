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

// IsAncestor checks whether ancestor is a git ancestor of descendant.
func (c *Client) IsAncestor(ctx context.Context, owner, repo, ancestor, descendant string) (bool, error) {
	result, err := c.v3Client.IsAncestor(ctx, owner, repo, ancestor, descendant)
	if err != nil {
		return false, fmt.Errorf("check ancestor: %w", err)
	}
	return result, nil
}
