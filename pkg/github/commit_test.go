//nolint:funlen
package github_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

func TestPullRequestCommit_ValidateUntrusted(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		commit                *github.PullRequestCommit
		trustedApps           map[string]struct{}
		trustedMachineUsers   map[string]struct{}
		untrustedMachineUsers map[string]struct{}
		want                  *github.UntrustedCommit
	}{
		{
			name: "trusted regular user with valid signature",
			commit: &github.PullRequestCommit{
				Commit: &github.Commit{
					OID: "abc123",
					Committer: &github.Committer{
						User: &github.User{
							Login:        "regular-user",
							ResourcePath: "/users/regular-user",
						},
					},
					Signature: &github.Signature{
						IsValid: true,
						State:   "valid",
					},
				},
			},
			want: nil,
		},
		{
			name: "not linked to user",
			commit: &github.PullRequestCommit{
				Commit: &github.Commit{
					OID: "abc123",
					Committer: &github.Committer{
						User: &github.User{},
					},
				},
			},
			want: &github.UntrustedCommit{
				NotLinkedToUser: true,
				SHA:             "abc123",
			},
		},
		{
			name: "invalid signature",
			commit: &github.PullRequestCommit{
				Commit: &github.Commit{
					OID: "abc123",
					Committer: &github.Committer{
						User: &github.User{
							Login:        "user",
							ResourcePath: "/users/user",
						},
					},
					Signature: &github.Signature{
						IsValid: false,
						State:   "invalid",
					},
				},
			},
			want: &github.UntrustedCommit{
				Login: "user",
				SHA:   "abc123",
				InvalidSign: &github.Signature{
					IsValid: false,
					State:   "invalid",
				},
			},
		},
		{
			name: "nil signature",
			commit: &github.PullRequestCommit{
				Commit: &github.Commit{
					OID: "abc123",
					Committer: &github.Committer{
						User: &github.User{
							Login:        "user",
							ResourcePath: "/users/user",
						},
					},
				},
			},
			want: &github.UntrustedCommit{
				Login:       "user",
				SHA:         "abc123",
				InvalidSign: nil,
			},
		},
		{
			name: "trusted app",
			commit: &github.PullRequestCommit{
				Commit: &github.Commit{
					OID: "abc123",
					Committer: &github.Committer{
						User: &github.User{
							Login:        "trusted-bot[bot]",
							ResourcePath: "/apps/trusted-bot",
						},
					},
					Signature: &github.Signature{
						IsValid: true,
						State:   "valid",
					},
				},
			},
			trustedApps: map[string]struct{}{
				"trusted-bot[bot]": {},
			},
			want: nil,
		},
		{
			name: "untrusted app",
			commit: &github.PullRequestCommit{
				Commit: &github.Commit{
					OID: "abc123",
					Committer: &github.Committer{
						User: &github.User{
							Login:        "untrusted-bot[bot]",
							ResourcePath: "/apps/untrusted-bot",
						},
					},
					Signature: &github.Signature{
						IsValid: true,
						State:   "valid",
					},
				},
			},
			want: &github.UntrustedCommit{
				Login:          "untrusted-bot[bot]",
				SHA:            "abc123",
				IsUntrustedApp: true,
			},
		},
		{
			name: "trusted machine user",
			commit: &github.PullRequestCommit{
				Commit: &github.Commit{
					OID: "abc123",
					Committer: &github.Committer{
						User: &github.User{
							Login:        "trusted-bot",
							ResourcePath: "/users/trusted-bot",
						},
					},
					Signature: &github.Signature{
						IsValid: true,
						State:   "valid",
					},
				},
			},
			trustedMachineUsers: map[string]struct{}{
				"trusted-bot": {},
			},
			want: nil,
		},
		{
			name: "untrusted machine user",
			commit: &github.PullRequestCommit{
				Commit: &github.Commit{
					OID: "abc123",
					Committer: &github.Committer{
						User: &github.User{
							Login:        "untrusted-bot",
							ResourcePath: "/users/untrusted-bot",
						},
					},
					Signature: &github.Signature{
						IsValid: true,
						State:   "valid",
					},
				},
			},
			untrustedMachineUsers: map[string]struct{}{
				"untrusted-*": {},
			},
			want: &github.UntrustedCommit{
				Login:                  "untrusted-bot",
				SHA:                    "abc123",
				IsUntrustedMachineUser: true,
			},
		},
		{
			name: "app takes precedence over machine user settings",
			commit: &github.PullRequestCommit{
				Commit: &github.Commit{
					OID: "abc123",
					Committer: &github.Committer{
						User: &github.User{
							Login:        "special-bot[bot]",
							ResourcePath: "/apps/special-bot",
						},
					},
					Signature: &github.Signature{
						IsValid: true,
						State:   "valid",
					},
				},
			},
			trustedMachineUsers: map[string]struct{}{
				"special-bot[bot]": {},
			},
			untrustedMachineUsers: map[string]struct{}{
				"special-*": {},
			},
			want: &github.UntrustedCommit{
				Login:          "special-bot[bot]",
				SHA:            "abc123",
				IsUntrustedApp: true,
			},
		},
		{
			name: "trusted machine user takes precedence over untrusted pattern",
			commit: &github.PullRequestCommit{
				Commit: &github.Commit{
					OID: "abc123",
					Committer: &github.Committer{
						User: &github.User{
							Login:        "automation-bot",
							ResourcePath: "/users/automation-bot",
						},
					},
					Signature: &github.Signature{
						IsValid: true,
						State:   "valid",
					},
				},
			},
			trustedMachineUsers: map[string]struct{}{
				"automation-bot": {},
			},
			untrustedMachineUsers: map[string]struct{}{
				"automation-*": {},
			},
			want: nil,
		},
		{
			name: "fallback to author when committer user is nil",
			commit: &github.PullRequestCommit{
				Commit: &github.Commit{
					OID: "abc123",
					Committer: &github.Committer{
						User: nil,
					},
					Author: &github.Committer{
						User: &github.User{
							Login:        "author-user",
							ResourcePath: "/users/author-user",
						},
					},
					Signature: &github.Signature{
						IsValid: true,
						State:   "valid",
					},
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.commit.ValidateUntrusted(tt.trustedApps, tt.trustedMachineUsers, tt.untrustedMachineUsers)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("PullRequestCommit.ValidateUntrusted() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCommit_User(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		commit *github.Commit
		want   *github.User
	}{
		{
			name: "committer user exists",
			commit: &github.Commit{
				Committer: &github.Committer{
					User: &github.User{
						Login:        "committer",
						ResourcePath: "/users/committer",
					},
				},
				Author: &github.Committer{
					User: &github.User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
				},
			},
			want: &github.User{
				Login:        "committer",
				ResourcePath: "/users/committer",
			},
		},
		{
			name: "committer user is nil, fallback to author",
			commit: &github.Commit{
				Committer: &github.Committer{
					User: nil,
				},
				Author: &github.Committer{
					User: &github.User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
				},
			},
			want: &github.User{
				Login:        "author",
				ResourcePath: "/users/author",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.commit.User()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Commit.User() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCommit_Linked(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		commit *github.Commit
		want   bool
	}{
		{
			name: "linked user with login",
			commit: &github.Commit{
				Committer: &github.Committer{
					User: &github.User{
						Login:        "user",
						ResourcePath: "/users/user",
					},
				},
			},
			want: true,
		},
		{
			name: "not linked user with empty login",
			commit: &github.Commit{
				Committer: &github.Committer{
					User: &github.User{
						Login:        "",
						ResourcePath: "/users/unknown",
					},
				},
			},
			want: false,
		},
		{
			name: "fallback to author when committer user is nil",
			commit: &github.Commit{
				Committer: &github.Committer{
					User: nil,
				},
				Author: &github.Committer{
					User: &github.User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
				},
			},
			want: true,
		},
		{
			name: "fallback to author with empty login",
			commit: &github.Commit{
				Committer: &github.Committer{
					User: nil,
				},
				Author: &github.Committer{
					User: &github.User{
						Login:        "",
						ResourcePath: "/users/unknown",
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.commit.Linked()
			if got != tt.want {
				t.Errorf("Commit.Linked() = %v, want %v", got, tt.want)
			}
		})
	}
}
