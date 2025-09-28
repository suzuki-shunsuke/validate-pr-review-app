package v4

import "github.com/shurcooL/githubv4"

type Review struct {
	Author    *User             `json:"author"`
	State     string            `json:"state"`
	Commit    *ReviewCommit     `json:"commit"`
	CreatedAt githubv4.DateTime `json:"createdAt"`
}

type ReviewCommit struct {
	OID string `json:"oid"`
}

type ListReviewsQuery struct {
	Repository *ReviewsRepository `graphql:"repository(owner: $repoOwner, name: $repoName)"`
}

func (q *ListReviewsQuery) Nodes() []*Review {
	return q.Repository.PullRequest.Reviews.Nodes
}

func (q *ListReviewsQuery) PageInfo() *PageInfo {
	return q.Repository.PullRequest.Reviews.PageInfo
}
