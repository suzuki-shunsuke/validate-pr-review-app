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
			name: "two approvals from different users",
			input: &validation.Input{
				Config: &config.Config{
					TrustedApps:                 []string{"trusted-bot"},
					UntrustedMachineUsers:       []string{"untrusted-bot"},
					TrustedMachineUsers:         []string{"trusted-user"},
					UniqueTrustedApps:           map[string]struct{}{"trusted-bot": {}},
					UniqueUntrustedMachineUsers: map[string]struct{}{"untrusted-bot": {}},
					UniqueTrustedMachineUsers:   map[string]struct{}{"trusted-user": {}},
				},
				PR: &github.PullRequest{
					HeadRefOID: "def456",
					Author: &github.User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{
							{
								Author: &github.User{
									Login:        "reviewer1",
									ResourcePath: "/users/reviewer1",
								},
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "def456",
								},
							},
							{
								Author: &github.User{
									Login:        "reviewer2",
									ResourcePath: "/users/reviewer2",
								},
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "def456",
								},
							},
						},
					},
					Commits: &github.Commits{
						Nodes: []*github.PullRequestCommit{
							{
								Commit: &github.Commit{
									OID: "def456",
									Committer: &github.Committer{
										User: &github.User{
											Login:        "author",
											ResourcePath: "/users/author",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State:                 validation.StateTwoApprovals,
				Approvers:             []string{"reviewer1", "reviewer2"},
				TrustedApps:           []string{"trusted-bot"},
				UntrustedMachineUsers: []string{"untrusted-bot"},
				TrustedMachineUsers:   []string{"trusted-user"},
			},
		},
		{
			name: "no approvals",
			input: &validation.Input{
				Config: &config.Config{
					TrustedApps:                 []string{},
					UntrustedMachineUsers:       []string{},
					TrustedMachineUsers:         []string{},
					UniqueTrustedApps:           map[string]struct{}{},
					UniqueUntrustedMachineUsers: map[string]struct{}{},
					UniqueTrustedMachineUsers:   map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadRefOID: "abc123",
					Author: &github.User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{},
					},
					Commits: &github.Commits{
						Nodes: []*github.PullRequestCommit{},
					},
				},
			},
			expected: &validation.Result{
				State:                 validation.StateApprovalIsRequired,
				TrustedApps:           []string{},
				UntrustedMachineUsers: []string{},
				TrustedMachineUsers:   []string{},
			},
		},
		{
			name: "approval_from_app_ignored",
			input: &validation.Input{
				Config: &config.Config{
					TrustedApps:                 []string{},
					UntrustedMachineUsers:       []string{},
					TrustedMachineUsers:         []string{},
					UniqueTrustedApps:           map[string]struct{}{},
					UniqueUntrustedMachineUsers: map[string]struct{}{},
					UniqueTrustedMachineUsers:   map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadRefOID: "def456",
					Author: &github.User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{
							{
								Author: &github.User{
									Login:        "github-bot[bot]",
									ResourcePath: "/apps/github-bot",
								},
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "def456",
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
				State:                 validation.StateApprovalIsRequired,
				TrustedApps:           []string{},
				UntrustedMachineUsers: []string{},
				TrustedMachineUsers:   []string{},
				IgnoredApprovers:      []string{"github-bot[bot]"},
			},
		},
		{
			name: "approval_from_untrusted_machine_user_ignored",
			input: &validation.Input{
				Config: &config.Config{
					TrustedApps:                 []string{},
					UntrustedMachineUsers:       []string{"untrusted-bot"},
					TrustedMachineUsers:         []string{},
					UniqueTrustedApps:           map[string]struct{}{},
					UniqueUntrustedMachineUsers: map[string]struct{}{"untrusted-bot": {}},
					UniqueTrustedMachineUsers:   map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadRefOID: "def456",
					Author: &github.User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{
							{
								Author: &github.User{
									Login:        "untrusted-bot",
									ResourcePath: "/users/untrusted-bot",
								},
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "def456",
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
				State:                 validation.StateApprovalIsRequired,
				TrustedApps:           []string{},
				UntrustedMachineUsers: []string{"untrusted-bot"},
				TrustedMachineUsers:   []string{},
				IgnoredApprovers:      []string{"untrusted-bot"},
			},
		},
		{
			name: "untrusted_author_requires_two_approvals",
			input: &validation.Input{
				Config: &config.Config{
					TrustedApps:                 []string{},
					UntrustedMachineUsers:       []string{"author"},
					TrustedMachineUsers:         []string{},
					UniqueTrustedApps:           map[string]struct{}{},
					UniqueUntrustedMachineUsers: map[string]struct{}{"author": {}},
					UniqueTrustedMachineUsers:   map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadRefOID: "def456",
					Author: &github.User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{
							{
								Author: &github.User{
									Login:        "reviewer1",
									ResourcePath: "/users/reviewer1",
								},
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "def456",
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
				State: validation.StateTwoApprovalsAreRequired,
				Author: &validation.User{
					Login: "author",
				},
				TrustedApps:           []string{},
				UntrustedMachineUsers: []string{"author"},
				TrustedMachineUsers:   []string{},
			},
		},
		{
			name: "one_approval_sufficient",
			input: &validation.Input{
				Config: &config.Config{
					TrustedApps:                 []string{},
					UntrustedMachineUsers:       []string{},
					TrustedMachineUsers:         []string{},
					UniqueTrustedApps:           map[string]struct{}{},
					UniqueUntrustedMachineUsers: map[string]struct{}{},
					UniqueTrustedMachineUsers:   map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadRefOID: "def456",
					Author: &github.User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
					Reviews: &github.Reviews{
						Nodes: []*github.Review{
							{
								Author: &github.User{
									Login:        "reviewer1",
									ResourcePath: "/users/reviewer1",
								},
								State: "APPROVED",
								Commit: &github.ReviewCommit{
									OID: "def456",
								},
							},
						},
					},
					Commits: &github.Commits{
						Nodes: []*github.PullRequestCommit{
							{
								Commit: &github.Commit{
									OID: "def456",
									Committer: &github.Committer{
										User: &github.User{
											Login:        "author",
											ResourcePath: "/users/author",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &validation.Result{
				Author: &validation.User{
					Login: "author",
				},
				Approvers:             []string{"author"},
				TrustedApps:           []string{},
				UntrustedMachineUsers: []string{},
				TrustedMachineUsers:   []string{},
			},
		},
	}

	ctrl := &validation.Controller{}

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
