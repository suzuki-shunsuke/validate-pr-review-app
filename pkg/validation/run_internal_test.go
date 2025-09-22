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
			name:     "user with empty login requires two approvals",
			user:     &github.User{},
			config:   nil,
			expected: true,
		},
		{
			name: "trusted app does not require two approvals",
			user: &github.User{
				Login:        "trusted-bot[bot]",
				ResourcePath: "/apps/trusted-bot",
			},
			config: &config.Config{
				UniqueTrustedApps: map[string]struct{}{"trusted-bot[bot]": {}},
			},
			expected: false,
		},
		{
			name: "untrusted app requires two approvals",
			user: &github.User{
				Login:        "untrusted-bot[bot]",
				ResourcePath: "/apps/untrusted-bot",
			},
			config:   &config.Config{},
			expected: true,
		},
		{
			name: "untrusted machine user requires two approvals",
			user: &github.User{
				Login:        "untrusted-user",
				ResourcePath: "/users/untrusted-user",
			},
			config: &config.Config{
				UniqueUntrustedMachineUsers: map[string]struct{}{"untrusted-user": {}},
			},
			expected: true,
		},
		{
			name: "regular user does not require two approvals",
			user: &github.User{
				Login:        "regular-user",
				ResourcePath: "/users/regular-user",
			},
			config:   &config.Config{},
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
			name: "approved review for head commit",
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
			name: "approved review for old commit",
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
			name: "non approved review",
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
			name: "dismissed review",
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
