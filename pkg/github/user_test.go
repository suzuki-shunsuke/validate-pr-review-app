//nolint:funlen
package github_test

import (
	"testing"

	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

func TestUser_IsApp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		user *github.User
		want bool
	}{
		{
			name: "both conditions met",
			user: &github.User{
				Login:        "dependabot[bot]",
				ResourcePath: "/apps/dependabot",
			},
			want: true,
		},
		{
			name: "neither condition met",
			user: &github.User{
				Login:        "regular-user",
				ResourcePath: "/users/regular-user",
			},
			want: false,
		},
		{
			name: "empty login and resource path",
			user: &github.User{
				Login:        "",
				ResourcePath: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.user.IsApp(); got != tt.want {
				t.Errorf("User.IsApp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsTrustedUser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		user                  *github.User
		trustedMachineUsers   map[string]struct{}
		untrustedMachineUsers map[string]struct{}
		want                  bool
	}{
		{
			name: "user in trusted machine users list",
			user: &github.User{Login: "trusted-bot"},
			trustedMachineUsers: map[string]struct{}{
				"trusted-bot": {},
			},
			want: true,
		},
		{
			name: "user matches untrusted pattern",
			user: &github.User{Login: "untrusted-bot"},
			untrustedMachineUsers: map[string]struct{}{
				"untrusted-*": {},
			},
			want: false,
		},
		{
			name: "user matches exact untrusted pattern",
			user: &github.User{Login: "specific-bot"},
			untrustedMachineUsers: map[string]struct{}{
				"specific-bot": {},
			},
			want: false,
		},
		{
			name: "regular user not in any list",
			user: &github.User{Login: "regular-user"},
			trustedMachineUsers: map[string]struct{}{
				"trusted-bot": {},
			},
			untrustedMachineUsers: map[string]struct{}{
				"untrusted-*": {},
			},
			want: true,
		},
		{
			name: "user in both lists - trusted takes precedence",
			user: &github.User{Login: "bot-user"},
			trustedMachineUsers: map[string]struct{}{
				"bot-user": {},
			},
			untrustedMachineUsers: map[string]struct{}{
				"bot-*": {},
			},
			want: true,
		},
		{
			name:                  "nil maps",
			user:                  &github.User{Login: "any-user"},
			trustedMachineUsers:   nil,
			untrustedMachineUsers: nil,
			want:                  true,
		},
		{
			name:                "invalid pattern in untrusted users",
			user:                &github.User{Login: "test-user"},
			trustedMachineUsers: map[string]struct{}{},
			untrustedMachineUsers: map[string]struct{}{
				"[": {}, // invalid pattern
			},
			want: true, // should continue and return true since no valid pattern matches
		},
		{
			name:                "multiple patterns, one matches",
			user:                &github.User{Login: "automation-bot"},
			trustedMachineUsers: map[string]struct{}{},
			untrustedMachineUsers: map[string]struct{}{
				"deploy-*":     {},
				"automation-*": {},
				"build-*":      {},
			},
			want: false,
		},
		{
			name:                "multiple patterns, none match",
			user:                &github.User{Login: "regular-user"},
			trustedMachineUsers: map[string]struct{}{},
			untrustedMachineUsers: map[string]struct{}{
				"deploy-*":     {},
				"automation-*": {},
				"build-*":      {},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.user.IsTrustedUser(tt.trustedMachineUsers, tt.untrustedMachineUsers); got != tt.want {
				t.Errorf("User.IsTrustedUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
