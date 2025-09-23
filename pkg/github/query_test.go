package github_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

func TestListCommitsQuery_PageInfo(t *testing.T) {
	t.Parallel()
	query := &github.ListCommitsQuery{
		Repository: &github.CommitsRepository{
			PullRequest: &github.CommitsPullRequest{
				Commits: &github.Commits{
					PageInfo: &github.PageInfo{
						HasNextPage: true,
						EndCursor:   "cursor123",
					},
				},
			},
		},
	}

	got := query.PageInfo()
	want := &github.PageInfo{
		HasNextPage: true,
		EndCursor:   "cursor123",
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ListCommitsQuery.PageInfo() mismatch (-want +got):\n%s", diff)
	}
}

func TestListCommitsQuery_Nodes(t *testing.T) {
	t.Parallel()
	nodes := []*github.PullRequestCommit{
		{
			Commit: &github.Commit{
				OID: "abc123",
			},
		},
		{
			Commit: &github.Commit{
				OID: "def456",
			},
		},
	}

	query := &github.ListCommitsQuery{
		Repository: &github.CommitsRepository{
			PullRequest: &github.CommitsPullRequest{
				Commits: &github.Commits{
					Nodes: nodes,
				},
			},
		},
	}

	got := query.Nodes()
	if diff := cmp.Diff(nodes, got); diff != "" {
		t.Errorf("ListCommitsQuery.Nodes() mismatch (-want +got):\n%s", diff)
	}
}

func TestListReviewsQuery_PageInfo(t *testing.T) {
	t.Parallel()
	query := &github.ListReviewsQuery{
		Repository: &github.ReviewsRepository{
			PullRequest: &github.ReviewsPullRequest{
				Reviews: &github.Reviews{
					PageInfo: &github.PageInfo{
						HasNextPage: false,
						EndCursor:   "cursor456",
					},
				},
			},
		},
	}

	got := query.PageInfo()
	want := &github.PageInfo{
		HasNextPage: false,
		EndCursor:   "cursor456",
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ListReviewsQuery.PageInfo() mismatch (-want +got):\n%s", diff)
	}
}

func TestListReviewsQuery_Nodes(t *testing.T) {
	t.Parallel()
	nodes := []*github.Review{
		{
			State: "APPROVED",
			Author: &github.User{
				Login: "reviewer1",
			},
		},
		{
			State: "CHANGES_REQUESTED",
			Author: &github.User{
				Login: "reviewer2",
			},
		},
	}

	query := &github.ListReviewsQuery{
		Repository: &github.ReviewsRepository{
			PullRequest: &github.ReviewsPullRequest{
				Reviews: &github.Reviews{
					Nodes: nodes,
				},
			},
		},
	}

	got := query.Nodes()
	if diff := cmp.Diff(nodes, got); diff != "" {
		t.Errorf("ListReviewsQuery.Nodes() mismatch (-want +got):\n%s", diff)
	}
}

func TestCommitter_Login(t *testing.T) {
	t.Parallel()
	committer := &github.Committer{
		User: &github.User{
			Login:        "test-user",
			ResourcePath: "/users/test-user",
		},
	}

	got := committer.Login()
	want := "test-user"

	if got != want {
		t.Errorf("Committer.Login() = %v, want %v", got, want)
	}
}
