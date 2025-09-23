package github

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestListCommitsQuery_PageInfo(t *testing.T) {
	t.Parallel()
	query := &ListCommitsQuery{
		Repository: &CommitsRepository{
			PullRequest: &CommitsPullRequest{
				Commits: &Commits{
					PageInfo: &PageInfo{
						HasNextPage: true,
						EndCursor:   "cursor123",
					},
				},
			},
		},
	}

	got := query.PageInfo()
	want := &PageInfo{
		HasNextPage: true,
		EndCursor:   "cursor123",
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ListCommitsQuery.PageInfo() mismatch (-want +got):\n%s", diff)
	}
}

func TestListCommitsQuery_Nodes(t *testing.T) {
	t.Parallel()
	nodes := []*PullRequestCommit{
		{
			Commit: &Commit{
				OID: "abc123",
			},
		},
		{
			Commit: &Commit{
				OID: "def456",
			},
		},
	}

	query := &ListCommitsQuery{
		Repository: &CommitsRepository{
			PullRequest: &CommitsPullRequest{
				Commits: &Commits{
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
	query := &ListReviewsQuery{
		Repository: &ReviewsRepository{
			PullRequest: &ReviewsPullRequest{
				Reviews: &Reviews{
					PageInfo: &PageInfo{
						HasNextPage: false,
						EndCursor:   "cursor456",
					},
				},
			},
		},
	}

	got := query.PageInfo()
	want := &PageInfo{
		HasNextPage: false,
		EndCursor:   "cursor456",
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ListReviewsQuery.PageInfo() mismatch (-want +got):\n%s", diff)
	}
}

func TestListReviewsQuery_Nodes(t *testing.T) {
	t.Parallel()
	nodes := []*Review{
		{
			State: "APPROVED",
			Author: &User{
				Login: "reviewer1",
			},
		},
		{
			State: "CHANGES_REQUESTED",
			Author: &User{
				Login: "reviewer2",
			},
		},
	}

	query := &ListReviewsQuery{
		Repository: &ReviewsRepository{
			PullRequest: &ReviewsPullRequest{
				Reviews: &Reviews{
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
	committer := &Committer{
		User: &User{
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
