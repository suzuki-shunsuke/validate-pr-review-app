//nolint:funlen,maintidx
package validation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/validation"
)

func TestController_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    *validation.Input
		expected *validation.Result
	}{
		{
			name: "two approvals - sufficient",
			input: &validation.Input{
				Config: &config.Config{},
				PR: &github.PullRequest{
					HeadRefOID: "abc123",
					Author: &github.User{
						Login:        "pr-author",
						ResourcePath: "/users/pr-author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{
							{
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "abc123",
								},
								Author: &github.User{
									Login:        "reviewer1",
									ResourcePath: "/users/reviewer1",
								},
							},
							{
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "abc123",
								},
								Author: &github.User{
									Login:        "reviewer2",
									ResourcePath: "/users/reviewer2",
								},
							},
						},
					},
					Commits: &github.Commits{
						Nodes: []*github.PullRequestCommit{
							{
								Commit: &github.Commit{
									OID: "abc123",
									Committer: &github.Committer{
										User: &github.User{
											Login:        "committer",
											ResourcePath: "/users/committer",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State:     validation.StateTwoApprovals,
				Approvers: []string{"reviewer1", "reviewer2"},
			},
		},
		{
			name: "no approvals - approval required",
			input: &validation.Input{
				Config: &config.Config{},
				PR: &github.PullRequest{
					HeadRefOID: "abc123",
					Author: &github.User{
						Login:        "pr-author",
						ResourcePath: "/users/pr-author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{},
					},
					Commits: &github.Commits{
						Nodes: []*github.PullRequestCommit{
							{
								Commit: &github.Commit{
									OID: "abc123",
									Committer: &github.Committer{
										User: &github.User{
											Login:        "committer",
											ResourcePath: "/users/committer",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State:            validation.StateApprovalIsRequired,
				IgnoredApprovers: map[string]*github.IgnoredApproval{},
			},
		},
		{
			name: "one approval with untrusted author - two approvals required",
			input: &validation.Input{
				Config: &config.Config{
					UntrustedMachineUsers:       []string{"untrusted-*"},
					UniqueUntrustedMachineUsers: map[string]struct{}{"untrusted-*": {}},
				},
				PR: &github.PullRequest{
					HeadRefOID: "abc123",
					Author: &github.User{
						Login:        "untrusted-author",
						ResourcePath: "/users/untrusted-author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{
							{
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "abc123",
								},
								Author: &github.User{
									Login:        "reviewer1",
									ResourcePath: "/users/reviewer1",
								},
							},
						},
					},
					Commits: &github.Commits{
						Nodes: []*github.PullRequestCommit{
							{
								Commit: &github.Commit{
									OID: "abc123",
									Committer: &github.Committer{
										User: &github.User{
											Login:        "committer",
											ResourcePath: "/users/committer",
										},
									},
									Signature: &github.Signature{
										IsValid: true,
										State:   "valid",
									},
								},
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State: validation.StateTwoApprovalsAreRequired,
				Author: &github.Author{
					Login:                "untrusted-author",
					UntrustedMachineUser: true,
				},
				IgnoredApprovers:      map[string]*github.IgnoredApproval{},
				UntrustedMachineUsers: []string{"untrusted-*"},
			},
		},
		{
			name: "one approval with self approval - two approvals required",
			input: &validation.Input{
				Config: &config.Config{},
				PR: &github.PullRequest{
					HeadRefOID: "abc123",
					Author: &github.User{
						Login:        "pr-author",
						ResourcePath: "/users/pr-author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{
							{
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "abc123",
								},
								Author: &github.User{
									Login:        "committer",
									ResourcePath: "/users/committer",
								},
							},
						},
					},
					Commits: &github.Commits{
						Nodes: []*github.PullRequestCommit{
							{
								Commit: &github.Commit{
									OID: "abc123",
									Committer: &github.Committer{
										User: &github.User{
											Login:        "committer",
											ResourcePath: "/users/committer",
										},
									},
									Signature: &github.Signature{
										IsValid: true,
										State:   "valid",
									},
								},
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State:            validation.StateTwoApprovalsAreRequired,
				SelfApprover:     "committer",
				IgnoredApprovers: map[string]*github.IgnoredApproval{},
			},
		},
		{
			name: "one approval from trusted user - sufficient",
			input: &validation.Input{
				Config: &config.Config{
					TrustedApps:                 []string{},
					TrustedMachineUsers:         []string{},
					UntrustedMachineUsers:       []string{},
					UniqueTrustedApps:           map[string]struct{}{},
					UniqueTrustedMachineUsers:   map[string]struct{}{},
					UniqueUntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadRefOID: "abc123",
					Author: &github.User{
						Login:        "pr-author",
						ResourcePath: "/users/pr-author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{
							{
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "abc123",
								},
								Author: &github.User{
									Login:        "reviewer1",
									ResourcePath: "/users/reviewer1",
								},
							},
						},
					},
					Commits: &github.Commits{
						Nodes: []*github.PullRequestCommit{
							{
								Commit: &github.Commit{
									OID: "abc123",
									Committer: &github.Committer{
										User: &github.User{
											Login:        "different-committer",
											ResourcePath: "/users/different-committer",
										},
									},
									Signature: &github.Signature{
										IsValid: true,
										State:   "valid",
									},
								},
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State:                 validation.StateOneApproval,
				Approvers:             []string{"reviewer1"},
				IgnoredApprovers:      map[string]*github.IgnoredApproval{},
				TrustedApps:           []string{},
				UntrustedMachineUsers: []string{},
				TrustedMachineUsers:   []string{},
			},
		},
		{
			name: "ignored app approval",
			input: &validation.Input{
				Config: &config.Config{
					TrustedApps:                 []string{},
					TrustedMachineUsers:         []string{},
					UntrustedMachineUsers:       []string{},
					UniqueTrustedApps:           map[string]struct{}{},
					UniqueTrustedMachineUsers:   map[string]struct{}{},
					UniqueUntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadRefOID: "abc123",
					Author: &github.User{
						Login:        "pr-author",
						ResourcePath: "/users/pr-author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{
							{
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "abc123",
								},
								Author: &github.User{
									Login:        "bot-app[bot]",
									ResourcePath: "/apps/bot-app",
								},
							},
						},
					},
					Commits: &github.Commits{
						Nodes: []*github.PullRequestCommit{},
					},
				},
			},
			expected: &validation.Result{
				State: validation.StateApprovalIsRequired,
				IgnoredApprovers: map[string]*github.IgnoredApproval{
					"bot-app[bot]": {
						Login: "bot-app[bot]",
						IsApp: true,
					},
				},
				TrustedApps:           []string{},
				UntrustedMachineUsers: []string{},
				TrustedMachineUsers:   []string{},
			},
		},
		{
			name: "one approval with untrusted commit - two approvals required",
			input: &validation.Input{
				Config: &config.Config{
					TrustedApps:                 []string{},
					TrustedMachineUsers:         []string{},
					UntrustedMachineUsers:       []string{},
					UniqueTrustedApps:           map[string]struct{}{},
					UniqueTrustedMachineUsers:   map[string]struct{}{},
					UniqueUntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadRefOID: "abc123",
					Author: &github.User{
						Login:        "pr-author",
						ResourcePath: "/users/pr-author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{
							{
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "abc123",
								},
								Author: &github.User{
									Login:        "reviewer1",
									ResourcePath: "/users/reviewer1",
								},
							},
						},
					},
					Commits: &github.Commits{
						Nodes: []*github.PullRequestCommit{
							{
								Commit: &github.Commit{
									OID: "abc123",
									Committer: &github.Committer{
										User: &github.User{
											Login:        "committer",
											ResourcePath: "/users/committer",
										},
									},
									Signature: &github.Signature{
										IsValid: false,
										State:   "invalid",
									},
								},
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State: validation.StateTwoApprovalsAreRequired,
				UntrustedCommits: []*github.UntrustedCommit{
					{
						Login: "committer",
						SHA:   "abc123",
						InvalidSign: &github.Signature{
							IsValid: false,
							State:   "invalid",
						},
					},
				},
				IgnoredApprovers:      map[string]*github.IgnoredApproval{},
				TrustedApps:           []string{},
				UntrustedMachineUsers: []string{},
				TrustedMachineUsers:   []string{},
			},
		},
	}

	ctrl := validation.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ctrl.Run(nil, tt.input)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("validation result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
