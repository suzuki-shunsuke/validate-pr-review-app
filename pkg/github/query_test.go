package github

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPullRequest_ValidateAuthor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		pr                    *PullRequest
		trustedApps           map[string]struct{}
		trustedMachineUsers   map[string]struct{}
		untrustedMachineUsers map[string]struct{}
		want                  *Author
	}{
		{
			name: "trusted regular user",
			pr: &PullRequest{
				Author: &User{
					Login:        "regular-user",
					ResourcePath: "/users/regular-user",
				},
			},
			want: nil,
		},
		{
			name: "trusted app",
			pr: &PullRequest{
				Author: &User{
					Login:        "trusted-bot[bot]",
					ResourcePath: "/apps/trusted-bot",
				},
			},
			trustedApps: map[string]struct{}{
				"trusted-bot[bot]": {},
			},
			want: nil,
		},
		{
			name: "untrusted app",
			pr: &PullRequest{
				Author: &User{
					Login:        "untrusted-bot[bot]",
					ResourcePath: "/apps/untrusted-bot",
				},
			},
			want: &Author{
				Login:        "untrusted-bot[bot]",
				UntrustedApp: true,
			},
		},
		{
			name: "trusted machine user",
			pr: &PullRequest{
				Author: &User{
					Login:        "automation-bot",
					ResourcePath: "/users/automation-bot",
				},
			},
			trustedMachineUsers: map[string]struct{}{
				"automation-bot": {},
			},
			want: nil,
		},
		{
			name: "untrusted machine user",
			pr: &PullRequest{
				Author: &User{
					Login:        "untrusted-bot",
					ResourcePath: "/users/untrusted-bot",
				},
			},
			untrustedMachineUsers: map[string]struct{}{
				"untrusted-*": {},
			},
			want: &Author{
				Login:                "untrusted-bot",
				UntrustedMachineUser: true,
			},
		},
		{
			name: "app takes precedence over machine user settings",
			pr: &PullRequest{
				Author: &User{
					Login:        "special-bot[bot]",
					ResourcePath: "/apps/special-bot",
				},
			},
			trustedMachineUsers: map[string]struct{}{
				"special-bot[bot]": {},
			},
			untrustedMachineUsers: map[string]struct{}{
				"special-*": {},
			},
			want: &Author{
				Login:        "special-bot[bot]",
				UntrustedApp: true,
			},
		},
		{
			name: "trusted machine user takes precedence over untrusted pattern",
			pr: &PullRequest{
				Author: &User{
					Login:        "automation-bot",
					ResourcePath: "/users/automation-bot",
				},
			},
			trustedMachineUsers: map[string]struct{}{
				"automation-bot": {},
			},
			untrustedMachineUsers: map[string]struct{}{
				"automation-*": {},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.pr.ValidateAuthor(tt.trustedApps, tt.trustedMachineUsers, tt.untrustedMachineUsers)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("PullRequest.ValidateAuthor() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

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