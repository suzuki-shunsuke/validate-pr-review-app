package v4

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

// GetPR gets a pull request reviews and committers via GitHub GraphQL API.
func (c *Client) GetPR(ctx context.Context, owner, name string, number int) (*PullRequest, error) {
	q := &GetPRQuery{}
	variables := map[string]any{
		"repoOwner": githubv4.String(owner),
		"repoName":  githubv4.String(name),
		"number":    githubv4.Int(number),
	}
	if err := c.v4Client.Query(ctx, q, variables); err != nil {
		return nil, fmt.Errorf("get a pull request by GitHub GraphQL API: %w", err)
	}

	// TODO Exclude reviews not associated with the latest commit
	if q.Repository.PullRequest.Reviews.PageInfo.HasNextPage {
		for range 10 {
			reviews, err := c.ListReviews(ctx, owner, name, number, q.Repository.PullRequest.Reviews.PageInfo.EndCursor)
			if err != nil {
				return nil, fmt.Errorf("list reviews by GitHub GraphQL API: %w", err)
			}
			q.Repository.PullRequest.Reviews.Nodes = append(q.Repository.PullRequest.Reviews.Nodes, reviews...)
		}
	}
	if q.Repository.PullRequest.Commits.PageInfo.HasNextPage {
		for range 10 {
			commits, err := c.ListCommits(ctx, owner, name, number, q.Repository.PullRequest.Commits.PageInfo.EndCursor)
			if err != nil {
				return nil, fmt.Errorf("list commits by GitHub GraphQL API: %w", err)
			}
			q.Repository.PullRequest.Commits.Nodes = append(q.Repository.PullRequest.Commits.Nodes, commits...)
		}
	}
	return q.Repository.PullRequest, nil
}

// ListReviews lists reviews of a pull request via GitHub GraphQL API.
func (c *Client) ListReviews(ctx context.Context, owner, name string, number int, cursor string) ([]*Review, error) {
	var reviews []*Review
	variables := map[string]any{
		"repoOwner": githubv4.String(owner),
		"repoName":  githubv4.String(name),
		"number":    githubv4.Int(number),
		"cursor":    githubv4.String(cursor),
	}
	for range 100 {
		q := &ListReviewsQuery{}
		if err := c.v4Client.Query(ctx, q, variables); err != nil {
			return nil, fmt.Errorf("list reviews by GitHub GraphQL API: %w", err)
		}
		reviews = append(reviews, q.Nodes()...)
		pageInfo := q.PageInfo()
		if !pageInfo.HasNextPage {
			return reviews, nil
		}
		variables["cursor"] = pageInfo.EndCursor
	}
	return reviews, nil
}

// ListCommits lists commits of a pull request via GitHub GraphQL API.
func (c *Client) ListCommits(ctx context.Context, owner, name string, number int, cursor string) ([]*PullRequestCommit, error) {
	var commits []*PullRequestCommit
	variables := map[string]any{
		"repoOwner": githubv4.String(owner),
		"repoName":  githubv4.String(name),
		"number":    githubv4.Int(number),
		"cursor":    githubv4.String(cursor),
	}
	for range 100 {
		q := &ListCommitsQuery{}
		if err := c.v4Client.Query(ctx, q, variables); err != nil {
			return nil, fmt.Errorf("list commits by GitHub GraphQL API: %w", err)
		}
		commits = append(commits, q.Nodes()...)
		pageInfo := q.PageInfo()
		if !pageInfo.HasNextPage {
			return commits, nil
		}
		variables["cursor"] = pageInfo.EndCursor
	}
	return commits, nil
}
