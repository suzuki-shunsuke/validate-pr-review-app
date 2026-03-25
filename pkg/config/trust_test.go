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
				TrustedApps: []string{"renovate[bot]", "dependabot[bot]"},
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
