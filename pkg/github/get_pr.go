package github

import (
	"context"

	v4 "github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github/v4"
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
		commits[i] = newCommit(v)
	}

	reviewsByCommit := groupReviewsByCommit(pr.Reviews.Nodes)
	approversByCommit := buildApproversByCommit(reviewsByCommit)

	p := &PullRequest{
		HeadSHA:           pr.HeadRefOID,
		BaseSHA:           pr.BaseRefOID,
		Commits:           commits,
		Approvers:         approversByCommit[pr.HeadRefOID],
		ApproversByCommit: approversByCommit,
	}
	if p.Approvers == nil {
		p.Approvers = make(map[string]*User)
	}
	return p, nil
}

// groupReviewsByCommit groups reviews by commit OID,
// keeping only the latest review per user per commit.
func groupReviewsByCommit(nodes []*v4.Review) map[string]map[string]*v4.Review {
	reviewsByCommit := make(map[string]map[string]*v4.Review)
	for _, node := range nodes {
		review := newReview(node)
		login := review.Author.Login
		if login == "" {
			continue
		}
		commitOID := node.Commit.OID
		if reviewsByCommit[commitOID] == nil {
			reviewsByCommit[commitOID] = make(map[string]*v4.Review)
		}
		if a, ok := reviewsByCommit[commitOID][login]; ok {
			if node.CreatedAt.Before(a.CreatedAt.Time) {
				continue
			}
		}
		reviewsByCommit[commitOID][login] = node
	}
	return reviewsByCommit
}

// buildApproversByCommit converts grouped reviews to approvers maps,
// filtering to only APPROVED reviews.
func buildApproversByCommit(reviewsByCommit map[string]map[string]*v4.Review) map[string]map[string]*User {
	approversByCommit := make(map[string]map[string]*User, len(reviewsByCommit))
	for oid, reviews := range reviewsByCommit {
		m := make(map[string]*User)
		for k, v := range reviews {
			if v.State == "APPROVED" {
				m[k] = newUser(v.Author)
			}
		}
		if len(m) > 0 {
			approversByCommit[oid] = m
		}
	}
	return approversByCommit
}
