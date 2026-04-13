package github

import "context"

// CompareCommits returns the list of changed file paths between two commits.
func (c *Client) CompareCommits(ctx context.Context, owner, repo, base, head string) ([]string, error) {
	return c.v3Client.CompareCommits(ctx, owner, repo, base, head)
}
