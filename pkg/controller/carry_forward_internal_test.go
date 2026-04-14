package controller

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
)

func Test_findCarryForwardApprovers(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name string
		pr   *github.PullRequest
		mock *mockGitHub
		want map[string]*github.User
	}{
		{
			name: "HEAD is empty commit, previous commit has review",
			pr: &github.PullRequest{
				HeadSHA: "empty1",
				Commits: []*github.Commit{
					{
						SHA:       "reviewed1",
						Committer: &github.User{Login: "alice"},
						Parents:   []string{"p0"},
					},
					{
						SHA:                     "empty1",
						Committer:               &github.User{Login: "bob"},
						Parents:                 []string{"reviewed1"},
						ChangedFilesIfAvailable: new(0),
					},
				},
				ApproversByCommit: map[string]map[string]*github.User{
					"reviewed1": {
						"carol": {Login: "carol"},
					},
				},
			},
			mock: &mockGitHub{},
			want: map[string]*github.User{
				"carol": {Login: "carol"},
			},
		},
		{
			name: "HEAD is clean merge commit, previous commit has review",
			pr: &github.PullRequest{
				HeadSHA: "merge1",
				Commits: []*github.Commit{
					{
						SHA:       "reviewed1",
						Committer: &github.User{Login: "alice"},
						Parents:   []string{"p0"},
					},
					{
						SHA:       "merge1",
						Committer: &github.User{Login: "bob"},
						Parents:   []string{"reviewed1", "main-tip"},
					},
				},
				ApproversByCommit: map[string]map[string]*github.User{
					"reviewed1": {
						"carol": {Login: "carol"},
					},
				},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"reviewed1...merge1": {"file_a.go"},
					"main-tip...merge1":  {"file_b.go"},
				},
			},
			want: map[string]*github.User{
				"carol": {Login: "carol"},
			},
		},
		{
			name: "HEAD is regular commit (non-empty, non-merge)",
			pr: &github.PullRequest{
				HeadSHA: "regular1",
				Commits: []*github.Commit{
					{
						SHA:       "reviewed1",
						Committer: &github.User{Login: "alice"},
						Parents:   []string{"p0"},
					},
					{
						SHA:                     "regular1",
						Committer:               &github.User{Login: "bob"},
						Parents:                 []string{"reviewed1"},
						ChangedFilesIfAvailable: new(5),
					},
				},
				ApproversByCommit: map[string]map[string]*github.User{
					"reviewed1": {
						"carol": {Login: "carol"},
					},
				},
			},
			mock: &mockGitHub{},
			want: nil,
		},
		{
			name: "HEAD is empty, no commit has reviews",
			pr: &github.PullRequest{
				HeadSHA: "empty1",
				Commits: []*github.Commit{
					{
						SHA:                     "commit1",
						Committer:               &github.User{Login: "alice"},
						Parents:                 []string{"p0"},
						ChangedFilesIfAvailable: new(0),
					},
					{
						SHA:                     "empty1",
						Committer:               &github.User{Login: "bob"},
						Parents:                 []string{"commit1"},
						ChangedFilesIfAvailable: new(0),
					},
				},
				ApproversByCommit: map[string]map[string]*github.User{},
			},
			mock: &mockGitHub{},
			want: nil,
		},
		{
			name: "multiple empty/merge commits walk back to reviewed commit",
			pr: &github.PullRequest{
				HeadSHA: "empty3",
				Commits: []*github.Commit{
					{
						SHA:       "reviewed1",
						Committer: &github.User{Login: "alice"},
						Parents:   []string{"p0"},
					},
					{
						SHA:                     "empty1",
						Committer:               &github.User{Login: "bob"},
						Parents:                 []string{"reviewed1"},
						ChangedFilesIfAvailable: new(0),
					},
					{
						SHA:                     "empty2",
						Committer:               &github.User{Login: "bob"},
						Parents:                 []string{"empty1"},
						ChangedFilesIfAvailable: new(0),
					},
					{
						SHA:                     "empty3",
						Committer:               &github.User{Login: "bob"},
						Parents:                 []string{"empty2"},
						ChangedFilesIfAvailable: new(0),
					},
				},
				ApproversByCommit: map[string]map[string]*github.User{
					"reviewed1": {
						"carol": {Login: "carol"},
					},
				},
			},
			mock: &mockGitHub{},
			want: map[string]*github.User{
				"carol": {Login: "carol"},
			},
		},
		{
			name: "conflict-resolution merge commit blocks carry-forward",
			pr: &github.PullRequest{
				HeadSHA: "conflict-merge1",
				Commits: []*github.Commit{
					{
						SHA:       "reviewed1",
						Committer: &github.User{Login: "alice"},
						Parents:   []string{"p0"},
					},
					{
						SHA:       "conflict-merge1",
						Committer: &github.User{Login: "bob"},
						Parents:   []string{"reviewed1", "main-tip"},
					},
				},
				ApproversByCommit: map[string]map[string]*github.User{
					"reviewed1": {
						"carol": {Login: "carol"},
					},
				},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"reviewed1...conflict-merge1": {"file_a.go", "file_b.go"},
					"main-tip...conflict-merge1":  {"file_b.go", "file_c.go"}, // overlap on file_b.go
				},
			},
			want: nil,
		},
		{
			name: "nil changedFilesIfAvailable with single parent blocks carry-forward",
			pr: &github.PullRequest{
				HeadSHA: "unknown1",
				Commits: []*github.Commit{
					{
						SHA:       "reviewed1",
						Committer: &github.User{Login: "alice"},
						Parents:   []string{"p0"},
					},
					{
						SHA:                     "unknown1",
						Committer:               &github.User{Login: "bob"},
						Parents:                 []string{"reviewed1"},
						ChangedFilesIfAvailable: nil,
					},
				},
				ApproversByCommit: map[string]map[string]*github.User{
					"reviewed1": {
						"carol": {Login: "carol"},
					},
				},
			},
			mock: &mockGitHub{},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := &Controller{gh: tt.mock}
			ev := &Event{RepoOwner: "owner", RepoName: "repo"}
			got := ctrl.findCarryForwardApprovers(context.Background(), discardLogger, ev, tt.pr)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("findCarryForwardApprovers() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
