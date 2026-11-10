package secret_test

import (
	"testing"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/secret"
)

func TestSecret_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		secret  *secret.Secret
		useKMS  bool
		wantErr bool
	}{
		{
			name: "private key",
			secret: &secret.Secret{
				GitHubAppPrivateKey: "private key",
				WebhookSecret:       "webhook secret",
			},
		},
		{
			name: "kms",
			secret: &secret.Secret{
				WebhookSecret: "webhook secret",
			},
			useKMS: true,
		},
		{
			name: "neither private key nor kms",
			secret: &secret.Secret{
				WebhookSecret: "webhook secret",
			},
			wantErr: true,
		},
		{
			name: "both private key and kms",
			secret: &secret.Secret{
				GitHubAppPrivateKey: "private key",
				WebhookSecret:       "webhook secret",
			},
			useKMS:  true,
			wantErr: true,
		},
		{
			name: "no webhook secret",
			secret: &secret.Secret{
				GitHubAppPrivateKey: "private key",
			},
			wantErr: true,
		},
		{
			name:    "nil",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.secret.Validate(tt.useKMS)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
