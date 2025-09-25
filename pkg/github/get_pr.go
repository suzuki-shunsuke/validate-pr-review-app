package github

import (
	"context"

	v4 "github.com/suzuki-shunsuke/require-pr-review-app/pkg/github/v4"
)

type Signature = v4.Signature

// GetPR gets a pull request reviews and committers via GitHub GraphQL API.
func (c *Client) GetPR(ctx context.Context, owner, name string, number int) (*PullRequest, error) {
	pr, err := c.v4Client.GetPR(ctx, owner, name, number)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	commits := make([]*Commit, len(pr.Commits.Nodes))
	for i, v := range pr.Commits.Nodes {
		commits[i] = &Commit{
			SHA:       v.Commit.OID,
			Committer: newUser(v.Commit.User()),
			Signature: v.Commit.Signature,
		}
	}
	// filter reviews
	// Get the latest review for each user
	reviews := make(map[string]*v4.Review, len(pr.Reviews.Nodes))
	for _, node := range pr.Reviews.Nodes {
		if node.Commit.OID != pr.HeadRefOID {
			// Exclude reviews for non head commits
			continue
		}
		review := newReview(node)
		login := review.Author.Login
		if login == "" {
			// Skip reviews from deleted users
			continue
		}
		if a, ok := reviews[login]; ok {
			// Keep the latest review
			if node.CreatedAt.Before(a.CreatedAt.Time) {
				continue
			}
			reviews[login] = node
			continue
		}
		reviews[login] = node
	}
	m := make(map[string]struct{}, len(reviews))
	for k, v := range reviews {
		if v.State != "APPROVED" {
			continue
		}
		m[k] = struct{}{}
	}
	p := &PullRequest{
		HeadSHA:   pr.HeadRefOID,
		Commits:   commits,
		Approvers: m,
	}
	return p, nil
}
