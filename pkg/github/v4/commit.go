package v4

type PullRequestCommit struct {
	Commit *Commit `json:"commit"`
}

type Commit struct {
	OID       string     `json:"oid"`
	Committer *Committer `json:"committer"`
	Author    *Committer `json:"author"`
	Signature *Signature `json:"signature"`
}

type Signature struct {
	IsValid bool   `json:"isValid"`
	State   string `json:"state"`
}

func (c *Commit) User() *User {
	if c.Committer.User != nil {
		return c.Committer.User
	}
	return c.Author.User
}

type ListCommitsQuery struct {
	Repository *CommitsRepository `graphql:"repository(owner: $repoOwner, name: $repoName)"`
}

func (q *ListCommitsQuery) PageInfo() *PageInfo {
	return q.Repository.PullRequest.Commits.PageInfo
}

func (q *ListCommitsQuery) Nodes() []*PullRequestCommit {
	return q.Repository.PullRequest.Commits.Nodes
}

type CommitsRepository struct {
	PullRequest *CommitsPullRequest `graphql:"pullRequest(number: $number)"`
}

type CommitsPullRequest struct {
	Commits *Commits `graphql:"commits(first:30, after:$cursor)"`
}
