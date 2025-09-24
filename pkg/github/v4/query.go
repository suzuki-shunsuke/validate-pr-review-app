package v4

/*
query($owner: String!, $repo: String!, $pr: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $pr) {
      headRefOid
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
	HeadRefOID string `json:"headRefOid"`
	// latestReviews isn't appropriate.
	// If someone adds a review comment after approval, the last review is the comment, not the approval.
	Reviews *Reviews `json:"reviews" graphql:"reviews(first:30, states: [APPROVED, DISMISSED, CHANGES_REQUESTED])"`
	Commits *Commits `json:"commits" graphql:"commits(first:30)"`
}

type Author struct {
	Login                string
	UntrustedMachineUser bool
	UntrustedApp         bool
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

type Commits struct {
	// TotalCount int                  `json:"totalCount"`
	PageInfo *PageInfo            `json:"pageInfo"`
	Nodes    []*PullRequestCommit `json:"nodes"`
}

type Committer struct {
	User *User `json:"user"`
}

func (c *Committer) Login() string {
	return c.User.Login
}
