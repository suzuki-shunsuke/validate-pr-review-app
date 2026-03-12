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
			name:     "two approvals - sufficient",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"reviewer1": {Login: "reviewer1"},
						"reviewer2": {Login: "reviewer2"},
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
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
			},
			expected: &validation.Result{
				State:     validation.StateApproved,
				Approvers: []string{"reviewer1", "reviewer2"},
			},
		},
		{
			name:     "no approvals - approval required",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadSHA:   "abc123",
					Approvers: map[string]*github.User{},
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
			name:     "one approval with self approval - two approvals required",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"committer": {Login: "committer"},
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
			name:     "one approval from trusted user - sufficient",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"reviewer1": {Login: "reviewer1"},
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
			name:     "ignored app approval",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"bot-app[bot]": {Login: "bot-app[bot]", IsApp: true},
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
			name:     "ignored app approval without bot suffix",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"coderabbitai": {Login: "coderabbitai", IsApp: true},
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
						Login: "coderabbitai",
						IsApp: true,
					},
				},
			},
		},
		{
			name:     "ignored untrusted machine user approval",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{"untrusted-*": {}},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"untrusted-bot": {Login: "untrusted-bot"},
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
			name:     "trusted machine user approval - sufficient",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{"trusted-bot": {}},
					UntrustedMachineUsers: map[string]struct{}{"trusted-*": {}},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"trusted-bot": {Login: "trusted-bot"},
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
			name:     "one approval with untrusted commit - two approvals required",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"reviewer1": {Login: "reviewer1"},
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
			name:     "one approval with untrusted app commit - two approvals required",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"reviewer1": {Login: "reviewer1"},
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
			name:     "unsigned commit with AllowUnsignedCommits - one approval sufficient",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				Insecure: &validation.Insecure{
					AllowUnsignedCommits: true,
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"reviewer1": {Login: "reviewer1"},
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
				State:     validation.StateApproved,
				Approvers: []string{"reviewer1"},
			},
		},
		{
			name:     "unsigned commit with matching UnsignedCommitAuthors - one approval sufficient",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				Insecure: &validation.Insecure{
					UnsignedCommitAuthors: []string{"machine-*"},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"reviewer1": {Login: "reviewer1"},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "machine-user",
								IsApp: false,
							},
							Signature: nil,
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
			name:     "unsigned commit with non-matching UnsignedCommitAuthors - two approvals required",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				Insecure: &validation.Insecure{
					UnsignedCommitAuthors: []string{"machine-*"},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"reviewer1": {Login: "reviewer1"},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "other-user",
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
						Login: "other-user",
						SHA:   "abc123",
						InvalidSign: &github.Signature{
							State: "invalid",
						},
					},
				},
			},
		},
		{
			name:     "unsigned commit with nil Insecure - two approvals required",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"reviewer1": {Login: "reviewer1"},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "committer",
								IsApp: false,
							},
							Signature: nil,
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
					},
				},
			},
		},
		{
			name:     "one approval with commit not linked to user - two approvals required",
			inputNew: &validation.InputNew{},
			input: &validation.Input{
				Trust: &validation.Trust{
					TrustedApps:           map[string]struct{}{},
					TrustedMachineUsers:   map[string]struct{}{},
					UntrustedMachineUsers: map[string]struct{}{},
				},
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]*github.User{
						"reviewer1": {Login: "reviewer1"},
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
