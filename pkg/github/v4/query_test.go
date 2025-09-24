package v4_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v4 "github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github/v4"
)

func TestListCommitsQuery_PageInfo(t *testing.T) {
	t.Parallel()
	query := &v4.ListCommitsQuery{
		Repository: &v4.CommitsRepository{
			PullRequest: &v4.CommitsPullRequest{
				Commits: &v4.Commits{
					PageInfo: &v4.PageInfo{
						HasNextPage: true,
						EndCursor:   "cursor123",
					},
				},
			},
		},
	}

	got := query.PageInfo()
	want := &v4.PageInfo{
		HasNextPage: true,
		EndCursor:   "cursor123",
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ListCommitsQuery.PageInfo() mismatch (-want +got):\n%s", diff)
	}
}

func TestListCommitsQuery_Nodes(t *testing.T) {
	t.Parallel()
	nodes := []*v4.PullRequestCommit{
		{
			Commit: &v4.Commit{
				OID: "abc123",
			},
		},
		{
			Commit: &v4.Commit{
				OID: "def456",
			},
		},
	}

	query := &v4.ListCommitsQuery{
		Repository: &v4.CommitsRepository{
			PullRequest: &v4.CommitsPullRequest{
				Commits: &v4.Commits{
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
	query := &v4.ListReviewsQuery{
		Repository: &v4.ReviewsRepository{
			PullRequest: &v4.ReviewsPullRequest{
				Reviews: &v4.Reviews{
					PageInfo: &v4.PageInfo{
						HasNextPage: false,
						EndCursor:   "cursor456",
					},
				},
			},
		},
	}

	got := query.PageInfo()
	want := &v4.PageInfo{
		HasNextPage: false,
		EndCursor:   "cursor456",
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ListReviewsQuery.PageInfo() mismatch (-want +got):\n%s", diff)
	}
}

func TestListReviewsQuery_Nodes(t *testing.T) {
	t.Parallel()
	nodes := []*v4.Review{
		{
			State: "APPROVED",
			Author: &v4.User{
				Login: "reviewer1",
			},
		},
		{
			State: "CHANGES_REQUESTED",
			Author: &v4.User{
				Login: "reviewer2",
			},
		},
	}

	query := &v4.ListReviewsQuery{
		Repository: &v4.ReviewsRepository{
			PullRequest: &v4.ReviewsPullRequest{
				Reviews: &v4.Reviews{
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
	committer := &v4.Committer{
		User: &v4.User{
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
