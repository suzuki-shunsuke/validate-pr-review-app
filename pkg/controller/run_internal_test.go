package controller

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
)

func Test_mergeTrust(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name   string
		global *config.Trust
		repo   *config.Trust
		want   config.Trust
	}{
		{
			name: "both nil",
			want: config.Trust{},
		},
		{
			name: "global set, repo nil",
			global: &config.Trust{
				TrustedApps:           []string{"app1[bot]"},
				
				UntrustedMachineUsers: []string{"evil*"},
			},
			want: config.Trust{
				TrustedApps:           []string{"app1[bot]"},
				
				UntrustedMachineUsers: []string{"evil*"},
			},
		},
		{
			name: "global set, repo partial override",
			global: &config.Trust{
				TrustedApps:           []string{"app1[bot]"},
				
				UntrustedMachineUsers: []string{"evil*"},
			},
			repo: &config.Trust{
				TrustedApps: []string{"app2[bot]"},
			},
			want: config.Trust{
				TrustedApps:           []string{"app2[bot]"},
				
				UntrustedMachineUsers: []string{"evil*"},
			},
		},
		{
			name: "global set, repo full override",
			global: &config.Trust{
				TrustedApps:           []string{"app1[bot]"},
				
				UntrustedMachineUsers: []string{"evil*"},
			},
			repo: &config.Trust{
				TrustedApps:           []string{"app2[bot]"},
				UntrustedMachineUsers: []string{"bad*"},
			},
			want: config.Trust{
				TrustedApps:           []string{"app2[bot]"},
				UntrustedMachineUsers: []string{"bad*"},
			},
		},
		{
			name:   "global nil, repo set",
			global: nil,
			repo: &config.Trust{
				TrustedApps: []string{"app2[bot]"},
			},
			want: config.Trust{
				TrustedApps: []string{"app2[bot]"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mergeTrust(tt.global, tt.repo)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mergeTrust() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_mergeTrust_doesNotMutateGlobal(t *testing.T) {
	t.Parallel()
	global := &config.Trust{
		TrustedApps:           []string{"app1[bot]"},
		
		UntrustedMachineUsers: []string{"evil*"},
	}
	repo := &config.Trust{
		TrustedApps:         []string{"app2[bot]"},
		
	}
	original := *global
	_ = mergeTrust(global, repo)
	if diff := cmp.Diff(original, *global); diff != "" {
		t.Errorf("mergeTrust mutated global (-before +after):\n%s", diff)
	}
}

func Test_mergeInsecure(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name   string
		global *config.Insecure
		repo   *config.Insecure
		want   config.Insecure
	}{
		{
			name: "both nil",
			want: config.Insecure{},
		},
		{
			name: "global set AllowUnsignedCommits true, repo nil",
			global: &config.Insecure{
				AllowUnsignedCommits: new(true),
			},
			want: config.Insecure{
				AllowUnsignedCommits: new(true),
			},
		},
		{
			name: "global set apps and machine users, repo nil",
			global: &config.Insecure{
				UnsignedCommitApps:         []string{"app1"},
				UnsignedCommitMachineUsers: []string{"bot1"},
			},
			want: config.Insecure{
				UnsignedCommitApps:         []string{"app1"},
				UnsignedCommitMachineUsers: []string{"bot1"},
			},
		},
		{
			name: "global set apps and machine users, repo overrides all",
			global: &config.Insecure{
				UnsignedCommitApps:         []string{"app1"},
				UnsignedCommitMachineUsers: []string{"bot1"},
			},
			repo: &config.Insecure{
				AllowUnsignedCommits:       new(false),
				UnsignedCommitApps:         []string{"app2"},
				UnsignedCommitMachineUsers: []string{"bot2"},
			},
			want: config.Insecure{
				AllowUnsignedCommits:       new(false),
				UnsignedCommitApps:         []string{"app2"},
				UnsignedCommitMachineUsers: []string{"bot2"},
			},
		},
		{
			name: "global set apps and machine users, repo partial override apps only",
			global: &config.Insecure{
				UnsignedCommitApps:         []string{"app1"},
				UnsignedCommitMachineUsers: []string{"bot1"},
			},
			repo: &config.Insecure{
				UnsignedCommitApps: []string{"app2"},
			},
			want: config.Insecure{
				AllowUnsignedCommits:       new(false),
				UnsignedCommitApps:         []string{"app2"},
				UnsignedCommitMachineUsers: []string{"bot1"},
			},
		},
		{
			name: "global set AllowUnsignedCommits true, repo sets apps",
			global: &config.Insecure{
				AllowUnsignedCommits: new(true),
			},
			repo: &config.Insecure{
				UnsignedCommitApps: []string{"app2"},
			},
			want: config.Insecure{
				AllowUnsignedCommits: new(false),
				UnsignedCommitApps:   []string{"app2"},
			},
		},
		{
			name: "global set apps, repo sets AllowUnsignedCommits true",
			global: &config.Insecure{
				UnsignedCommitApps:         []string{"app1"},
				UnsignedCommitMachineUsers: []string{"bot1"},
			},
			repo: &config.Insecure{
				AllowUnsignedCommits: new(true),
			},
			want: config.Insecure{
				AllowUnsignedCommits: new(true),
			},
		},
		{
			name: "global set apps and machine users, repo sets machine users only",
			global: &config.Insecure{
				UnsignedCommitApps:         []string{"app1"},
				UnsignedCommitMachineUsers: []string{"bot1"},
			},
			repo: &config.Insecure{
				UnsignedCommitMachineUsers: []string{"bot2"},
			},
			want: config.Insecure{
				AllowUnsignedCommits:       new(false),
				UnsignedCommitApps:         []string{"app1"},
				UnsignedCommitMachineUsers: []string{"bot2"},
			},
		},
		{
			name:   "global nil, repo set",
			global: nil,
			repo: &config.Insecure{
				AllowUnsignedCommits: new(true),
			},
			want: config.Insecure{
				AllowUnsignedCommits: new(true),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mergeInsecure(tt.global, tt.repo)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mergeInsecure() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_mergeInsecure_doesNotMutateGlobal(t *testing.T) {
	t.Parallel()
	global := &config.Insecure{
		UnsignedCommitApps:         []string{"app1"},
		UnsignedCommitMachineUsers: []string{"bot1"},
	}
	repo := &config.Insecure{
		AllowUnsignedCommits:       new(false),
		UnsignedCommitApps:         []string{"app2"},
		UnsignedCommitMachineUsers: []string{"bot2"},
	}
	original := *global
	_ = mergeInsecure(global, repo)
	if diff := cmp.Diff(original, *global); diff != "" {
		t.Errorf("mergeInsecure mutated global (-before +after):\n%s", diff)
	}
}
