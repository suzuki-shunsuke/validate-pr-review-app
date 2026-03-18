package config_test

import (
	"testing"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
)

func TestInsecure_Validate(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name    string
		input   *config.Insecure
		wantErr bool
	}{
		{
			name: "allow_unsigned_commits true only",
			input: &config.Insecure{
				AllowUnsignedCommits: new(true),
			},
		},
		{
			name: "unsigned_commit_apps only",
			input: &config.Insecure{
				UnsignedCommitApps: []string{"renovate[bot]"},
			},
		},
		{
			name: "unsigned_commit_machine_users only",
			input: &config.Insecure{
				UnsignedCommitMachineUsers: []string{"bot-user"},
			},
		},
		{
			name: "allow_unsigned_commits false with apps",
			input: &config.Insecure{
				AllowUnsignedCommits: new(false),
				UnsignedCommitApps:   []string{"renovate[bot]"},
			},
		},
		{
			name: "allow_unsigned_commits true with apps",
			input: &config.Insecure{
				AllowUnsignedCommits: new(true),
				UnsignedCommitApps:   []string{"renovate[bot]"},
			},
			wantErr: true,
		},
		{
			name: "allow_unsigned_commits true with machine users",
			input: &config.Insecure{
				AllowUnsignedCommits:       new(true),
				UnsignedCommitMachineUsers: []string{"bot-user"},
			},
			wantErr: true,
		},
		{
			name: "allow_unsigned_commits true with both apps and machine users",
			input: &config.Insecure{
				AllowUnsignedCommits:       new(true),
				UnsignedCommitApps:         []string{"renovate[bot]"},
				UnsignedCommitMachineUsers: []string{"bot-user"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Insecure.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
