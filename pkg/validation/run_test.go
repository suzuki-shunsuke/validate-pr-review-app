//nolint:funlen,maintidx
package validation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/validation"
)

func TestController_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		inputNew *validation.InputNew
		input    *validation.Input
		expected *validation.Result
	}{
		{
			name: "two approvals - sufficient",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						"reviewer1": {},
						"reviewer2": {},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "committer",
								IsApp: false,
							},
							Signature: &github.Signature{
								IsValid: true,
								State:   "valid",
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State:     validation.StateApproved,
				Approvers: []string{"reviewer1", "reviewer2"},
			},
		},
		{
			name: "no approvals - approval required",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA:   "abc123",
					Approvers: map[string]struct{}{},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "committer",
								IsApp: false,
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State: validation.StateApprovalIsRequired,
			},
		},
		{
			name: "one approval with self approval - two approvals required",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						"committer": {},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "committer",
								IsApp: false,
							},
							Signature: &github.Signature{
								IsValid: true,
								State:   "valid",
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State:        validation.StateTwoApprovalsAreRequired,
				SelfApprover: "committer",
			},
		},
		{
			name: "one approval from trusted user - sufficient",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						"reviewer1": {},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "different-committer",
								IsApp: false,
							},
							Signature: &github.Signature{
								IsValid: true,
								State:   "valid",
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State:     validation.StateApproved,
				Approvers: []string{"reviewer1"},
			},
		},
		{
			name: "ignored app approval",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						"bot-app[bot]": {},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "committer",
								IsApp: false,
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State: validation.StateApprovalIsRequired,
				IgnoredApprovers: []*github.IgnoredApproval{
					{
						Login: "bot-app[bot]",
						IsApp: true,
					},
				},
			},
		},
		{
			name: "ignored untrusted machine user approval",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{"untrusted-*": {}},
			},
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						"untrusted-bot": {},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "committer",
								IsApp: false,
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State: validation.StateApprovalIsRequired,
				IgnoredApprovers: []*github.IgnoredApproval{
					{
						Login:                  "untrusted-bot",
						IsUntrustedMachineUser: true,
					},
				},
			},
		},
		{
			name: "trusted machine user approval - sufficient",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{"trusted-bot": {}},
				UntrustedMachineUsers: map[string]struct{}{"trusted-*": {}},
			},
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						"trusted-bot": {},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "different-committer",
								IsApp: false,
							},
							Signature: &github.Signature{
								IsValid: true,
								State:   "valid",
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State:     validation.StateApproved,
				Approvers: []string{"trusted-bot"},
			},
		},
		{
			name: "one approval with untrusted commit - two approvals required",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						"reviewer1": {},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "committer",
								IsApp: false,
							},
							Signature: &github.Signature{
								IsValid: false,
								State:   "invalid",
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
							State: "invalid",
						},
					},
				},
			},
		},
		{
			name: "one approval with untrusted app commit - two approvals required",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						"reviewer1": {},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "untrusted-app[bot]",
								IsApp: true,
							},
							Signature: &github.Signature{
								IsValid: true,
								State:   "valid",
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State: validation.StateTwoApprovalsAreRequired,
				UntrustedCommits: []*github.UntrustedCommit{
					{
						Login:          "untrusted-app[bot]",
						SHA:            "abc123",
						IsUntrustedApp: true,
					},
				},
			},
		},
		{
			name: "one approval with commit not linked to user - two approvals required",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						"reviewer1": {},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "", // Empty login means not linked to user
								IsApp: false,
							},
						},
					},
				},
			},
			expected: &validation.Result{
				State: validation.StateTwoApprovalsAreRequired,
				UntrustedCommits: []*github.UntrustedCommit{
					{
						SHA:             "abc123",
						NotLinkedToUser: true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := validation.New(tt.inputNew)
			result := ctrl.Run(nil, tt.input)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("validation result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIsApp(t *testing.T) { //nolint:gocognit,cyclop
	t.Parallel()
	// Since isApp is not exported, we test it indirectly through the Run method
	tests := []struct {
		name     string
		approver string
		isApp    bool
	}{
		{
			name:     "regular user",
			approver: "regular-user",
			isApp:    false,
		},
		{
			name:     "bot user with [bot] suffix",
			approver: "dependabot[bot]",
			isApp:    true,
		},
		{
			name:     "app user with [bot] suffix",
			approver: "github-actions[bot]",
			isApp:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			inputNew := &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			}
			ctrl := validation.New(inputNew)

			input := &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						tt.approver: {},
					},
					Commits: []*github.Commit{},
				},
			}

			result := ctrl.Run(nil, input)

			if tt.isApp { //nolint:nestif
				// Should be in ignored approvers
				found := false
				for _, ignored := range result.IgnoredApprovers {
					if ignored.Login == tt.approver && ignored.IsApp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected %s to be detected as app and ignored", tt.approver)
				}
			} else {
				// Should be in regular approvers (if only one approver) or not in ignored
				if len(result.Approvers) == 1 && result.Approvers[0] == tt.approver {
					// Good, detected as regular user
					return
				}
				// Or should not be in ignored approvers as app
				for _, ignored := range result.IgnoredApprovers {
					if ignored.Login == tt.approver && ignored.IsApp {
						t.Errorf("Expected %s to NOT be detected as app", tt.approver)
					}
				}
			}
		})
	}
}

func TestController_VerifyUser(t *testing.T) { //nolint:gocognit,cyclop
	t.Parallel()
	tests := []struct {
		name            string
		inputNew        *validation.InputNew
		login           string
		expectedTrusted bool
	}{
		{
			name: "regular user",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
			login:           "regular-user",
			expectedTrusted: true,
		},
		{
			name: "trusted machine user",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{"trusted-bot": {}},
				UntrustedMachineUsers: map[string]struct{}{},
			},
			login:           "trusted-bot",
			expectedTrusted: true,
		},
		{
			name: "untrusted machine user by pattern",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{"untrusted-*": {}},
			},
			login:           "untrusted-bot",
			expectedTrusted: false,
		},
		{
			name: "trusted user takes precedence over untrusted pattern",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{"automation-bot": {}},
				UntrustedMachineUsers: map[string]struct{}{"automation-*": {}},
			},
			login:           "automation-bot",
			expectedTrusted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := validation.New(tt.inputNew)

			// Test indirectly through Run method
			input := &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						tt.login: {},
					},
					Commits: []*github.Commit{},
				},
			}

			result := ctrl.Run(nil, input)

			if tt.expectedTrusted { //nolint:nestif
				// Should be in approvers (if only one) or not marked as untrusted in ignored
				if len(result.Approvers) == 1 && result.Approvers[0] == tt.login {
					return // Good
				}
				// Check not in ignored as untrusted machine user
				for _, ignored := range result.IgnoredApprovers {
					if ignored.Login == tt.login && ignored.IsUntrustedMachineUser {
						t.Errorf("Expected %s to be trusted, but was marked as untrusted machine user", tt.login)
					}
				}
			} else {
				// Should be in ignored approvers as untrusted machine user
				found := false
				for _, ignored := range result.IgnoredApprovers {
					if ignored.Login == tt.login && ignored.IsUntrustedMachineUser {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected %s to be marked as untrusted machine user", tt.login)
				}
			}
		})
	}
}
