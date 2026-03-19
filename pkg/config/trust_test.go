package config_test

import (
	"testing"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
)

func TestTrust_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		trust   *config.Trust
		wantErr bool
	}{
		{
			name: "valid entries",
			trust: &config.Trust{
				TrustedApps:         []string{"renovate[bot]", "dependabot[bot]"},
				TrustedMachineUsers: []string{"bot-user", "ci-bot"},
			},
		},
		{
			name:  "empty trust",
			trust: &config.Trust{},
		},
		{
			name: "invalid trusted app with dot",
			trust: &config.Trust{
				TrustedApps: []string{"user.name"},
			},
			wantErr: true,
		},
		{
			name: "invalid trusted app with asterisk",
			trust: &config.Trust{
				TrustedApps: []string{"renovate*"},
			},
			wantErr: true,
		},
		{
			name: "invalid trusted machine user with dot",
			trust: &config.Trust{
				TrustedMachineUsers: []string{"user.name"},
			},
			wantErr: true,
		},
		{
			name: "invalid trusted machine user with asterisk",
			trust: &config.Trust{
				TrustedMachineUsers: []string{"bot*"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.trust.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Trust.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
