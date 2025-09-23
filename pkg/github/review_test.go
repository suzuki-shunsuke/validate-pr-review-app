package github

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReview_Ignored(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		review    *Review
		latestSHA string
		want      bool
	}{
		{
			name: "approved review for latest commit",
			review: &Review{
				State: "APPROVED",
				Commit: &ReviewCommit{
					OID: "abc123",
				},
			},
			latestSHA: "abc123",
			want:      false, // APPROVED and SHA matches = not ignored
		},
		{
			name: "approved review for old commit",
			review: &Review{
				State: "APPROVED",
				Commit: &ReviewCommit{
					OID: "old123",
				},
			},
			latestSHA: "abc123",
			want:      true, // APPROVED but SHA mismatch = ignored
		},
		{
			name: "changes requested for latest commit",
			review: &Review{
				State: "CHANGES_REQUESTED",
				Commit: &ReviewCommit{
					OID: "abc123",
				},
			},
			latestSHA: "abc123",
			want:      true, // Not APPROVED = ignored
		},
		{
			name: "changes requested for old commit",
			review: &Review{
				State: "CHANGES_REQUESTED",
				Commit: &ReviewCommit{
					OID: "old123",
				},
			},
			latestSHA: "abc123",
			want:      true, // Not APPROVED and SHA mismatch = ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.review.Ignored(tt.latestSHA); got != tt.want {
				t.Errorf("Review.Ignored() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReview_ValidateIgnored(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		review                *Review
		trustedMachineUsers   map[string]struct{}
		untrustedMachineUsers map[string]struct{}
		want                  *IgnoredApproval
	}{
		{
			name: "app user - ignored",
			review: &Review{
				Author: &User{
					Login:        "github-bot[bot]",
					ResourcePath: "/apps/github-bot",
				},
				State: "APPROVED",
			},
			want: &IgnoredApproval{
				Login: "github-bot[bot]",
				IsApp: true,
			},
		},
		{
			name: "trusted machine user - not ignored",
			review: &Review{
				Author: &User{
					Login:        "trusted-bot",
					ResourcePath: "/users/trusted-bot",
				},
				State: "APPROVED",
			},
			trustedMachineUsers: map[string]struct{}{
				"trusted-bot": {},
			},
			want: nil,
		},
		{
			name: "untrusted machine user - ignored",
			review: &Review{
				Author: &User{
					Login:        "untrusted-bot",
					ResourcePath: "/users/untrusted-bot",
				},
				State: "APPROVED",
			},
			untrustedMachineUsers: map[string]struct{}{
				"untrusted-*": {},
			},
			want: &IgnoredApproval{
				Login:                  "untrusted-bot",
				IsUntrustedMachineUser: true,
			},
		},
		{
			name: "regular user - not ignored",
			review: &Review{
				Author: &User{
					Login:        "regular-user",
					ResourcePath: "/users/regular-user",
				},
				State: "APPROVED",
			},
			want: nil,
		},
		{
			name: "app takes precedence over untrusted machine user",
			review: &Review{
				Author: &User{
					Login:        "bot-app[bot]",
					ResourcePath: "/apps/bot-app",
				},
				State: "APPROVED",
			},
			trustedMachineUsers: map[string]struct{}{},
			untrustedMachineUsers: map[string]struct{}{
				"bot-*": {},
			},
			want: &IgnoredApproval{
				Login: "bot-app[bot]",
				IsApp: true,
			},
		},
		{
			name: "trusted machine user takes precedence over untrusted pattern",
			review: &Review{
				Author: &User{
					Login:        "special-bot",
					ResourcePath: "/users/special-bot",
				},
				State: "APPROVED",
			},
			trustedMachineUsers: map[string]struct{}{
				"special-bot": {},
			},
			untrustedMachineUsers: map[string]struct{}{
				"*-bot": {},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.review.ValidateIgnored(tt.trustedMachineUsers, tt.untrustedMachineUsers)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Review.ValidateIgnored() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
