package config

import "testing"

func Test_validateLoginNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		names   []string
		field   string
		wantErr bool
	}{
		{
			name:  "valid names",
			names: []string{"renovate[bot]", "bot-user", "dependabot[bot]"},
			field: "trusted_apps",
		},
		{
			name:  "empty list",
			names: []string{},
			field: "trusted_apps",
		},
		{
			name:  "nil list",
			names: nil,
			field: "trusted_apps",
		},
		{
			name:    "invalid with dot",
			names:   []string{"user.name"},
			field:   "trusted_apps",
			wantErr: true,
		},
		{
			name:    "invalid with asterisk",
			names:   []string{"renovate*"},
			field:   "trusted_machine_users",
			wantErr: true,
		},
		{
			name:    "invalid with both",
			names:   []string{".*bot"},
			field:   "trusted_apps",
			wantErr: true,
		},
		{
			name:    "valid followed by invalid",
			names:   []string{"valid-name", "invalid.name"},
			field:   "trusted_apps",
			wantErr: true,
		},
		{
			name:    "invalid with question mark",
			names:   []string{"bot?"},
			field:   "trusted_apps",
			wantErr: true,
		},
		{
			name:    "invalid with caret",
			names:   []string{"^bot"},
			field:   "trusted_apps",
			wantErr: true,
		},
		{
			name:    "invalid with plus",
			names:   []string{"bot+"},
			field:   "trusted_apps",
			wantErr: true,
		},
		{
			name:    "invalid with dollar",
			names:   []string{"bot$"},
			field:   "trusted_apps",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateLoginNames(tt.names, tt.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLoginNames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
