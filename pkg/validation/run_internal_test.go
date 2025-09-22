//nolint:funlen
package validation

import (
	"testing"

	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

func Test_checkIfUserRequiresTwoApprovals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		user     *github.User
		config   *config.Config
		expected bool
	}{
		{
			name: "user_with_empty_login_requires_two_approvals",
			user: &github.User{
				Login:        "",
				ResourcePath: "/users/unknown",
			},
			config:   nil,
			expected: true,
		},
		{
			name: "trusted_app_does_not_require_two_approvals",
			user: &github.User{
				Login:        "trusted-bot[bot]",
				ResourcePath: "/apps/trusted-bot",
			},
			config: &config.Config{
				UniqueTrustedApps:           map[string]struct{}{"trusted-bot[bot]": {}},
				UniqueUntrustedMachineUsers: map[string]struct{}{},
			},
			expected: false,
		},
		{
			name: "untrusted_app_requires_two_approvals",
			user: &github.User{
				Login:        "untrusted-bot[bot]",
				ResourcePath: "/apps/untrusted-bot",
			},
			config: &config.Config{
				UniqueTrustedApps:           map[string]struct{}{},
				UniqueUntrustedMachineUsers: map[string]struct{}{},
			},
			expected: true,
		},
		{
			name: "untrusted_machine_user_requires_two_approvals",
			user: &github.User{
				Login:        "untrusted-user",
				ResourcePath: "/users/untrusted-user",
			},
			config: &config.Config{
				UniqueTrustedApps:           map[string]struct{}{},
				UniqueUntrustedMachineUsers: map[string]struct{}{"untrusted-user": {}},
			},
			expected: true,
		},
		{
			name: "regular_user_does_not_require_two_approvals",
			user: &github.User{
				Login:        "regular-user",
				ResourcePath: "/users/regular-user",
			},
			config: &config.Config{
				UniqueTrustedApps:           map[string]struct{}{},
				UniqueUntrustedMachineUsers: map[string]struct{}{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := &Input{Config: tt.config}
			result := checkIfUserRequiresTwoApprovals(tt.user, input)
			if result != tt.expected {
				t.Errorf("checkIfUserRequiresTwoApprovals() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func Test_isLatestApproval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		review     *github.Review
		headRefOID string
		expected   bool
	}{
		{
			name: "approved_review_for_head_commit",
			review: &github.Review{
				State: "APPROVED",
				Commit: &github.ReviewCommit{
					OID: "abc123",
				},
			},
			headRefOID: "abc123",
			expected:   true,
		},
		{
			name: "approved_review_for_old_commit",
			review: &github.Review{
				State: "APPROVED",
				Commit: &github.ReviewCommit{
					OID: "abc123",
				},
			},
			headRefOID: "def456",
			expected:   false,
		},
		{
			name: "non_approved_review",
			review: &github.Review{
				State: "CHANGES_REQUESTED",
				Commit: &github.ReviewCommit{
					OID: "abc123",
				},
			},
			headRefOID: "abc123",
			expected:   false,
		},
		{
			name: "dismissed_review",
			review: &github.Review{
				State: "DISMISSED",
				Commit: &github.ReviewCommit{
					OID: "abc123",
				},
			},
			headRefOID: "abc123",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isLatestApproval(tt.review, tt.headRefOID)
			if result != tt.expected {
				t.Errorf("isLatestApproval() = %v, want %v", result, tt.expected)
			}
		})
	}
}
