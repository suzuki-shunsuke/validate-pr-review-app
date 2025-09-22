package github

import "strings"

/*
query($owner: String!, $repo: String!, $pr: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $pr) {
      headRefOid
      author {
        login
        resourcePath
      }
      latestReviews(first: 100) {
        pageInfo {
          hasNextPage
          endCursor
        }
        nodes {
          state
          commit {
            oid
          }
          author {
            login
            resourcePath
          }
        }
      }
      commits(first: 100) {
        pageInfo {
          hasNextPage
          endCursor
        }
        nodes {
          commit {
            oid
            committer {
              user {
                login
                resourcePath
              }
            }
            author {
              user {
                login
                resourcePath
              }
            }
			signature {
			  isValid
			  state
			}
          }
        }
      }
    }
  }
}
*/

type PullRequest struct {
	HeadRefOID string   `json:"headRefOid"`
	Author     *User    `json:"author"`
	Reviews    *Reviews `json:"latestReviews" graphql:"latestReviews(first:30)"`
	Commits    *Commits `json:"commits" graphql:"commits(first:30)"`
}

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type GetPRQuery struct {
	Repository *Repository `graphql:"repository(owner: $repoOwner, name: $repoName)"`
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

type ListReviewsQuery struct {
	Repository *ReviewsRepository `graphql:"repository(owner: $repoOwner, name: $repoName)"`
}

func (q *ListReviewsQuery) Nodes() []*Review {
	return q.Repository.PullRequest.Reviews.Nodes
}

func (q *ListReviewsQuery) PageInfo() *PageInfo {
	return q.Repository.PullRequest.Reviews.PageInfo
}

type Repository struct {
	PullRequest *PullRequest `graphql:"pullRequest(number: $number)"`
}

type CommitsRepository struct {
	PullRequest *CommitsPullRequest `graphql:"pullRequest(number: $number)"`
}

type CommitsPullRequest struct {
	Commits *Commits `graphql:"commits(first:30)"`
}

type ReviewsRepository struct {
	PullRequest *ReviewsPullRequest `graphql:"pullRequest(number: $number)"`
}

type ReviewsPullRequest struct {
	Reviews *Reviews `graphql:"latestReviews(first:30)"`
}

type Reviews struct {
	// TotalCount int       `json:"totalCount"`
	PageInfo *PageInfo `json:"pageInfo"`
	Nodes    []*Review `json:"nodes"`
}

type Review struct {
	Author *User         `json:"author"`
	State  string        `json:"state"`
	Commit *ReviewCommit `json:"commit"`
}

type ReviewCommit struct {
	OID string `json:"oid"`
}

type Commits struct {
	// TotalCount int                  `json:"totalCount"`
	PageInfo *PageInfo            `json:"pageInfo"`
	Nodes    []*PullRequestCommit `json:"nodes"`
}

type PullRequestCommit struct {
	Commit *Commit `json:"commit"`
}

func (c *PullRequestCommit) Signature() *Signature {
	if c == nil || c.Commit == nil {
		return nil
	}
	return c.Commit.Signature
}

func (c *PullRequestCommit) User() *User {
	if c == nil || c.Commit == nil {
		return nil
	}
	return c.Commit.User()
}

func (c *PullRequestCommit) SHA() string {
	if c == nil || c.Commit == nil {
		return ""
	}
	return c.Commit.OID
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
	if c == nil {
		return nil
	}
	if user := c.Committer.GetUser(); user != nil {
		return user
	}
	return c.Author.GetUser()
}

func (c *Commit) Login() string {
	return c.User().GetLogin()
}

func (c *Commit) Linked() bool {
	return c.Login() != ""
}

type Committer struct {
	User *User `json:"user"`
}

func (c *Committer) GetUser() *User {
	if c == nil {
		return nil
	}
	return c.User
}

func (c *Committer) Login() string {
	if c == nil {
		return ""
	}
	return c.User.GetLogin()
}

type User struct {
	Login        string `json:"login"`
	ResourcePath string `json:"resourcePath"`
}

func (u *User) GetLogin() string {
	if u == nil {
		return ""
	}
	return u.Login
}

func (u *User) IsApp() bool {
	return strings.HasPrefix(u.ResourcePath, "/apps/") || strings.HasSuffix(u.Login, "[bot]")
}

func (u *User) Trusted(reliableBots map[string]struct{}) bool {
	_, ok := reliableBots[u.Login]
	return ok
}
