package controller

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
)

var discardLogger = slog.New(slog.DiscardHandler) //nolint:gochecknoglobals

type mockGitHub struct {
	compareResult  map[string][]string // key: "base...head"
	compareErr     map[string]error    // key: "base...head"
	ancestorResult map[string]bool     // key: "ancestor...descendant"
	ancestorErr    map[string]error    // key: "ancestor...descendant"
}

func (m *mockGitHub) GetPR(_ context.Context, _, _ string, _ int) (*github.PullRequest, error) {
	return nil, nil //nolint:nilnil
}

func (m *mockGitHub) CreateCheckRun(_ context.Context, _ githubv4.CreateCheckRunInput) error {
	return nil
}

func (m *mockGitHub) CompareCommits(_ context.Context, _, _, base, head string) ([]string, error) {
	key := base + "..." + head
	if err, ok := m.compareErr[key]; ok {
		return nil, err
	}
	if files, ok := m.compareResult[key]; ok {
		return files, nil
	}
	return nil, nil
}

func (m *mockGitHub) IsAncestor(_ context.Context, _, _, ancestor, descendant string) (bool, error) {
	key := ancestor + "..." + descendant
	if err, ok := m.ancestorErr[key]; ok {
		return false, err
	}
	if result, ok := m.ancestorResult[key]; ok {
		return result, nil
	}
	return false, nil
}

func Test_isCleanMergeCommit(t *testing.T) { //nolint:funlen
	t.Parallel()
	defaultPRCommitSHAs := map[string]struct{}{
		"parent1": {},
	}
	defaultBaseSHA := "base-sha"
	tests := []struct {
		name         string
		commit       *github.Commit
		mock         *mockGitHub
		prCommitSHAs map[string]struct{}
		baseSHA      string
		want         bool
	}{
		{
			name: "non-merge commit (single parent)",
			commit: &github.Commit{
				SHA:     "merge123",
				Parents: []string{"parent1"},
			},
			mock:         &mockGitHub{},
			prCommitSHAs: defaultPRCommitSHAs,
			baseSHA:      defaultBaseSHA,
			want:         false,
		},
		{
			name: "non-merge commit (no parents)",
			commit: &github.Commit{
				SHA:     "merge123",
				Parents: nil,
			},
			mock:         &mockGitHub{},
			prCommitSHAs: defaultPRCommitSHAs,
			baseSHA:      defaultBaseSHA,
			want:         false,
		},
		{
			name: "clean merge commit (no overlapping files, base branch parent)",
			commit: &github.Commit{
				SHA:     "merge123",
				Parents: []string{"parent1", "parent2"},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"parent1...merge123": {"file_a.go", "file_b.go"},
					"parent2...merge123": {"file_c.go", "file_d.go"},
				},
				ancestorResult: map[string]bool{
					"parent2...base-sha": true,
				},
			},
			prCommitSHAs: defaultPRCommitSHAs,
			baseSHA:      defaultBaseSHA,
			want:         true,
		},
		{
			name: "merge commit with overlapping files (conflict resolution)",
			commit: &github.Commit{
				SHA:     "merge123",
				Parents: []string{"parent1", "parent2"},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"parent1...merge123": {"file_a.go", "file_b.go"},
					"parent2...merge123": {"file_b.go", "file_c.go"},
				},
			},
			prCommitSHAs: defaultPRCommitSHAs,
			baseSHA:      defaultBaseSHA,
			want:         false,
		},
		{
			name: "compare API failure",
			commit: &github.Commit{
				SHA:     "merge123",
				Parents: []string{"parent1", "parent2"},
			},
			mock: &mockGitHub{
				compareErr: map[string]error{
					"parent1...merge123": errors.New("API error"),
				},
			},
			prCommitSHAs: defaultPRCommitSHAs,
			baseSHA:      defaultBaseSHA,
			want:         false,
		},
		{
			name: "too many changed files (>= 300)",
			commit: &github.Commit{
				SHA:     "merge123",
				Parents: []string{"parent1", "parent2"},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"parent1...merge123": make([]string, 300),
				},
			},
			prCommitSHAs: defaultPRCommitSHAs,
			baseSHA:      defaultBaseSHA,
			want:         false,
		},
		{
			name: "octopus merge (3 parents, no overlap, base branch parents)",
			commit: &github.Commit{
				SHA:     "merge123",
				Parents: []string{"parent1", "parent2", "parent3"},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"parent1...merge123": {"file_a.go"},
					"parent2...merge123": {"file_b.go"},
					"parent3...merge123": {"file_c.go"},
				},
				ancestorResult: map[string]bool{
					"parent2...base-sha": true,
					"parent3...base-sha": true,
				},
			},
			prCommitSHAs: defaultPRCommitSHAs,
			baseSHA:      defaultBaseSHA,
			want:         true,
		},
		{
			name: "octopus merge (3 parents, overlap between non-adjacent)",
			commit: &github.Commit{
				SHA:     "merge123",
				Parents: []string{"parent1", "parent2", "parent3"},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"parent1...merge123": {"file_a.go"},
					"parent2...merge123": {"file_b.go"},
					"parent3...merge123": {"file_a.go"},
				},
			},
			prCommitSHAs: defaultPRCommitSHAs,
			baseSHA:      defaultBaseSHA,
			want:         false,
		},
		{
			name: "non-PR parent is not ancestor of base branch",
			commit: &github.Commit{
				SHA:     "merge123",
				Parents: []string{"parent1", "arbitrary-branch"},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"parent1...merge123":          {"file_a.go"},
					"arbitrary-branch...merge123": {"file_b.go"},
				},
				ancestorResult: map[string]bool{
					"arbitrary-branch...base-sha": false,
				},
			},
			prCommitSHAs: defaultPRCommitSHAs,
			baseSHA:      defaultBaseSHA,
			want:         false,
		},
		{
			name: "ancestor check API failure",
			commit: &github.Commit{
				SHA:     "merge123",
				Parents: []string{"parent1", "parent2"},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"parent1...merge123": {"file_a.go"},
					"parent2...merge123": {"file_b.go"},
				},
				ancestorErr: map[string]error{
					"parent2...base-sha": errors.New("API error"),
				},
			},
			prCommitSHAs: defaultPRCommitSHAs,
			baseSHA:      defaultBaseSHA,
			want:         false,
		},
		{
			name: "all parents in PR commit list (no external parent)",
			commit: &github.Commit{
				SHA:     "merge123",
				Parents: []string{"parent1", "parent2"},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"parent1...merge123": {"file_a.go"},
					"parent2...merge123": {"file_b.go"},
				},
			},
			prCommitSHAs: map[string]struct{}{
				"parent1": {},
				"parent2": {},
			},
			baseSHA: defaultBaseSHA,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := &Controller{gh: tt.mock}
			ev := &Event{RepoOwner: "owner", RepoName: "repo"}
			got := ctrl.isCleanMergeCommit(context.Background(), discardLogger, ev, tt.commit, tt.prCommitSHAs, tt.baseSHA)
			if got != tt.want {
				t.Errorf("isCleanMergeCommit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkApproverCommits(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name                     string
		pr                       *github.PullRequest
		mock                     *mockGitHub
		wantIsAllowedMergeCommit []bool
	}{
		{
			name: "non-approver commits are skipped",
			pr: &github.PullRequest{
				Approvers: map[string]*github.User{
					"alice": {Login: "alice"},
				},
				Commits: []*github.Commit{
					{
						SHA:       "commit1",
						Committer: &github.User{Login: "bob"},
						Parents:   []string{"p1"},
					},
				},
			},
			mock:                     &mockGitHub{},
			wantIsAllowedMergeCommit: []bool{false},
		},
		{
			name: "approver clean merge commits marked as allowed",
			pr: &github.PullRequest{
				BaseSHA: "base-sha",
				Approvers: map[string]*github.User{
					"alice": {Login: "alice"},
				},
				Commits: []*github.Commit{
					{
						SHA:       "p1",
						Committer: &github.User{Login: "bob"},
						Parents:   []string{"p0"},
					},
					{
						SHA:       "merge1",
						Committer: &github.User{Login: "alice"},
						Parents:   []string{"p1", "p2"},
					},
					{
						SHA:       "merge2",
						Committer: &github.User{Login: "alice"},
						Parents:   []string{"merge1", "p4"},
					},
				},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"p1...merge1":     {"a.go"},
					"p2...merge1":     {"b.go"},
					"merge1...merge2": {"c.go"},
					"p4...merge2":     {"d.go"},
				},
				ancestorResult: map[string]bool{
					"p2...base-sha": true,
					"p4...base-sha": true,
				},
			},
			wantIsAllowedMergeCommit: []bool{false, true, true},
		},
		{
			name: "early termination on non-clean approver commit",
			pr: &github.PullRequest{
				Approvers: map[string]*github.User{
					"alice": {Login: "alice"},
				},
				Commits: []*github.Commit{
					{
						SHA:       "regular",
						Committer: &github.User{Login: "alice"},
						Parents:   []string{"p1"},
					},
					{
						SHA:       "merge1",
						Committer: &github.User{Login: "alice"},
						Parents:   []string{"p2", "p3"},
					},
				},
			},
			mock: &mockGitHub{
				compareResult: map[string][]string{
					"p2...merge1": {"a.go"},
					"p3...merge1": {"b.go"},
				},
			},
			// First commit is not a clean merge (single parent), so early termination.
			// Second commit is not checked, stays false (default).
			wantIsAllowedMergeCommit: []bool{false, false},
		},
		{
			name: "nil committer commit is skipped",
			pr: &github.PullRequest{
				Approvers: map[string]*github.User{
					"alice": {Login: "alice"},
				},
				Commits: []*github.Commit{
					{
						SHA:       "commit1",
						Committer: nil,
						Parents:   []string{"p1", "p2"},
					},
				},
			},
			mock:                     &mockGitHub{},
			wantIsAllowedMergeCommit: []bool{false},
		},
		{
			name: "approver empty commit (single parent) is allowed",
			pr: &github.PullRequest{
				Approvers: map[string]*github.User{
					"alice": {Login: "alice"},
				},
				Commits: []*github.Commit{
					{
						SHA:                     "empty1",
						Committer:               &github.User{Login: "alice"},
						Parents:                 []string{"p1"},
						ChangedFilesIfAvailable: new(0),
					},
				},
			},
			mock:                     &mockGitHub{},
			wantIsAllowedMergeCommit: []bool{true},
		},
		{
			name: "approver empty merge commit is allowed without Compare API",
			pr: &github.PullRequest{
				Approvers: map[string]*github.User{
					"alice": {Login: "alice"},
				},
				Commits: []*github.Commit{
					{
						SHA:                     "empty-merge1",
						Committer:               &github.User{Login: "alice"},
						Parents:                 []string{"p1", "p2"},
						ChangedFilesIfAvailable: new(0),
					},
				},
			},
			// No compareResult needed — empty commit check short-circuits before Compare API.
			mock:                     &mockGitHub{},
			wantIsAllowedMergeCommit: []bool{true},
		},
		{
			name: "approver commit with nil changedFilesIfAvailable is not allowed (fail closed)",
			pr: &github.PullRequest{
				Approvers: map[string]*github.User{
					"alice": {Login: "alice"},
				},
				Commits: []*github.Commit{
					{
						SHA:                     "commit1",
						Committer:               &github.User{Login: "alice"},
						Parents:                 []string{"p1"},
						ChangedFilesIfAvailable: nil,
					},
				},
			},
			mock:                     &mockGitHub{},
			wantIsAllowedMergeCommit: []bool{false},
		},
		{
			name: "approver non-empty single-parent commit triggers early termination",
			pr: &github.PullRequest{
				Approvers: map[string]*github.User{
					"alice": {Login: "alice"},
				},
				Commits: []*github.Commit{
					{
						SHA:                     "nonempty1",
						Committer:               &github.User{Login: "alice"},
						Parents:                 []string{"p1"},
						ChangedFilesIfAvailable: new(5),
					},
					{
						SHA:                     "empty1",
						Committer:               &github.User{Login: "alice"},
						Parents:                 []string{"p2"},
						ChangedFilesIfAvailable: new(0),
					},
				},
			},
			mock: &mockGitHub{},
			// First commit is non-empty single-parent → early termination.
			// Second commit is not checked.
			wantIsAllowedMergeCommit: []bool{false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := &Controller{gh: tt.mock}
			ev := &Event{RepoOwner: "owner", RepoName: "repo"}
			ctrl.checkApproverCommits(context.Background(), discardLogger, ev, tt.pr)

			got := make([]bool, len(tt.pr.Commits))
			for i, c := range tt.pr.Commits {
				got[i] = c.IsAllowedMergeCommit
			}
			if diff := cmp.Diff(tt.wantIsAllowedMergeCommit, got); diff != "" {
				t.Errorf("IsAllowedMergeCommit mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
