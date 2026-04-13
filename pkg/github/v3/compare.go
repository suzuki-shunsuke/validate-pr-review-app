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
