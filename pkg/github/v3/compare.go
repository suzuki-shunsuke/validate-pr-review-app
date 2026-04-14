package v3

import (
	"context"
	"fmt"
)

// CompareCommits compares two commits and returns the list of changed file paths.
func (c *Client) CompareCommits(ctx context.Context, owner, repo, base, head string) ([]string, error) {
	comp, _, err := c.client.Repositories.CompareCommits(ctx, owner, repo, base, head, nil)
	if err != nil {
		return nil, fmt.Errorf("compare commits %s...%s: %w", base, head, err)
	}
	files := make([]string, len(comp.Files))
	for i, f := range comp.Files {
		files[i] = f.GetFilename()
	}
	return files, nil
}

// IsAncestor checks whether ancestor is a git ancestor of descendant.
// It uses the Compare Two Commits API and checks that ancestor has
// no commits not reachable from descendant (BehindBy == 0).
func (c *Client) IsAncestor(ctx context.Context, owner, repo, ancestor, descendant string) (bool, error) {
	comp, _, err := c.client.Repositories.CompareCommits(ctx, owner, repo, ancestor, descendant, nil)
	if err != nil {
		return false, fmt.Errorf("compare commits %s...%s: %w", ancestor, descendant, err)
	}
	return comp.GetBehindBy() == 0, nil
}
