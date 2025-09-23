//nolint:funlen
package github

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPullRequestCommit_ValidateUntrusted(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		commit                *PullRequestCommit
		trustedApps           map[string]struct{}
		trustedMachineUsers   map[string]struct{}
		untrustedMachineUsers map[string]struct{}
		want                  *UntrustedCommit
	}{
		{
			name: "trusted regular user with valid signature",
			commit: &PullRequestCommit{
				Commit: &Commit{
					OID: "abc123",
					Committer: &Committer{
						User: &User{
							Login:        "regular-user",
							ResourcePath: "/users/regular-user",
						},
					},
					Signature: &Signature{
						IsValid: true,
						State:   "valid",
					},
				},
			},
			want: nil,
		},
		{
			name: "not linked to user",
			commit: &PullRequestCommit{
				Commit: &Commit{
					OID: "abc123",
					Committer: &Committer{
						User: &User{},
					},
				},
			},
			want: &UntrustedCommit{
				NotLinkedToUser: true,
				SHA:             "abc123",
			},
		},
		{
			name: "invalid signature",
			commit: &PullRequestCommit{
				Commit: &Commit{
					OID: "abc123",
					Committer: &Committer{
						User: &User{
							Login:        "user",
							ResourcePath: "/users/user",
						},
					},
					Signature: &Signature{
						IsValid: false,
						State:   "invalid",
					},
				},
			},
			want: &UntrustedCommit{
				Login: "user",
				SHA:   "abc123",
				InvalidSign: &Signature{
					IsValid: false,
					State:   "invalid",
				},
			},
		},
		{
			name: "nil signature",
			commit: &PullRequestCommit{
				Commit: &Commit{
					OID: "abc123",
					Committer: &Committer{
						User: &User{
							Login:        "user",
							ResourcePath: "/users/user",
						},
					},
				},
			},
			want: &UntrustedCommit{
				Login:       "user",
				SHA:         "abc123",
				InvalidSign: nil,
			},
		},
		{
			name: "trusted app",
			commit: &PullRequestCommit{
				Commit: &Commit{
					OID: "abc123",
					Committer: &Committer{
						User: &User{
							Login:        "trusted-bot[bot]",
							ResourcePath: "/apps/trusted-bot",
						},
					},
					Signature: &Signature{
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
			commit: &PullRequestCommit{
				Commit: &Commit{
					OID: "abc123",
					Committer: &Committer{
						User: &User{
							Login:        "untrusted-bot[bot]",
							ResourcePath: "/apps/untrusted-bot",
						},
					},
					Signature: &Signature{
						IsValid: true,
						State:   "valid",
					},
				},
			},
			want: &UntrustedCommit{
				Login:          "untrusted-bot[bot]",
				SHA:            "abc123",
				IsUntrustedApp: true,
			},
		},
		{
			name: "trusted machine user",
			commit: &PullRequestCommit{
				Commit: &Commit{
					OID: "abc123",
					Committer: &Committer{
						User: &User{
							Login:        "trusted-bot",
							ResourcePath: "/users/trusted-bot",
						},
					},
					Signature: &Signature{
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
			commit: &PullRequestCommit{
				Commit: &Commit{
					OID: "abc123",
					Committer: &Committer{
						User: &User{
							Login:        "untrusted-bot",
							ResourcePath: "/users/untrusted-bot",
						},
					},
					Signature: &Signature{
						IsValid: true,
						State:   "valid",
					},
				},
			},
			untrustedMachineUsers: map[string]struct{}{
				"untrusted-*": {},
			},
			want: &UntrustedCommit{
				Login:                  "untrusted-bot",
				SHA:                    "abc123",
				IsUntrustedMachineUser: true,
			},
		},
		{
			name: "app takes precedence over machine user settings",
			commit: &PullRequestCommit{
				Commit: &Commit{
					OID: "abc123",
					Committer: &Committer{
						User: &User{
							Login:        "special-bot[bot]",
							ResourcePath: "/apps/special-bot",
						},
					},
					Signature: &Signature{
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
			want: &UntrustedCommit{
				Login:          "special-bot[bot]",
				SHA:            "abc123",
				IsUntrustedApp: true,
			},
		},
		{
			name: "trusted machine user takes precedence over untrusted pattern",
			commit: &PullRequestCommit{
				Commit: &Commit{
					OID: "abc123",
					Committer: &Committer{
						User: &User{
							Login:        "automation-bot",
							ResourcePath: "/users/automation-bot",
						},
					},
					Signature: &Signature{
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
			commit: &PullRequestCommit{
				Commit: &Commit{
					OID: "abc123",
					Committer: &Committer{
						User: nil,
					},
					Author: &Committer{
						User: &User{
							Login:        "author-user",
							ResourcePath: "/users/author-user",
						},
					},
					Signature: &Signature{
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
		commit *Commit
		want   *User
	}{
		{
			name: "committer user exists",
			commit: &Commit{
				Committer: &Committer{
					User: &User{
						Login:        "committer",
						ResourcePath: "/users/committer",
					},
				},
				Author: &Committer{
					User: &User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
				},
			},
			want: &User{
				Login:        "committer",
				ResourcePath: "/users/committer",
			},
		},
		{
			name: "committer user is nil, fallback to author",
			commit: &Commit{
				Committer: &Committer{
					User: nil,
				},
				Author: &Committer{
					User: &User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
				},
			},
			want: &User{
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
		commit *Commit
		want   bool
	}{
		{
			name: "linked user with login",
			commit: &Commit{
				Committer: &Committer{
					User: &User{
						Login:        "user",
						ResourcePath: "/users/user",
					},
				},
			},
			want: true,
		},
		{
			name: "not linked user with empty login",
			commit: &Commit{
				Committer: &Committer{
					User: &User{
						Login:        "",
						ResourcePath: "/users/unknown",
					},
				},
			},
			want: false,
		},
		{
			name: "fallback to author when committer user is nil",
			commit: &Commit{
				Committer: &Committer{
					User: nil,
				},
				Author: &Committer{
					User: &User{
						Login:        "author",
						ResourcePath: "/users/author",
					},
				},
			},
			want: true,
		},
		{
			name: "fallback to author with empty login",
			commit: &Commit{
				Committer: &Committer{
					User: nil,
				},
				Author: &Committer{
					User: &User{
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
