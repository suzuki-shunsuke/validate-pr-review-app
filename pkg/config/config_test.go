package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConfig_Init(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                                string
		config                              *Config
		expectedUniqueTrustedApps           map[string]struct{}
		expectedUniqueTrustedMachineUsers   map[string]struct{}
		expectedUniqueUntrustedMachineUsers map[string]struct{}
		expectedCheckName                   string
		wantErr                             bool
	}{
		{
			name: "basic initialization",
			config: &Config{
				TrustedApps:           []string{"app1[bot]", "app2[bot]"},
				TrustedMachineUsers:   []string{"trusted-user1", "trusted-user2"},
				UntrustedMachineUsers: []string{"untrusted-*", "bot-*"},
				CheckName:             "custom-check",
				Templates:             map[string]string{},
			},
			expectedUniqueTrustedApps: map[string]struct{}{
				"app1[bot]": {},
				"app2[bot]": {},
			},
			expectedUniqueTrustedMachineUsers: map[string]struct{}{
				"trusted-user1": {},
				"trusted-user2": {},
			},
			expectedUniqueUntrustedMachineUsers: map[string]struct{}{
				"untrusted-*": {},
				"bot-*":       {},
			},
			expectedCheckName: "custom-check",
		},
		{
			name: "empty configuration with defaults",
			config: &Config{
				Templates: map[string]string{},
			},
			expectedUniqueTrustedApps:           map[string]struct{}{},
			expectedUniqueTrustedMachineUsers:   map[string]struct{}{},
			expectedUniqueUntrustedMachineUsers: map[string]struct{}{},
			expectedCheckName:                   "check-approval", // default value
		},
		{
			name: "duplicate entries in arrays",
			config: &Config{
				TrustedApps:           []string{"app1[bot]", "app1[bot]", "app2[bot]"},
				TrustedMachineUsers:   []string{"user1", "user1", "user2"},
				UntrustedMachineUsers: []string{"bot-*", "bot-*"},
				Templates:             map[string]string{},
			},
			expectedUniqueTrustedApps: map[string]struct{}{
				"app1[bot]": {},
				"app2[bot]": {},
			},
			expectedUniqueTrustedMachineUsers: map[string]struct{}{
				"user1": {},
				"user2": {},
			},
			expectedUniqueUntrustedMachineUsers: map[string]struct{}{
				"bot-*": {},
			},
			expectedCheckName: "check-approval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Init()

			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check unique trusted apps
			if diff := cmp.Diff(tt.expectedUniqueTrustedApps, tt.config.UniqueTrustedApps); diff != "" {
				t.Errorf("UniqueTrustedApps mismatch (-want +got):\n%s", diff)
			}

			// Check unique trusted machine users
			if diff := cmp.Diff(tt.expectedUniqueTrustedMachineUsers, tt.config.UniqueTrustedMachineUsers); diff != "" {
				t.Errorf("UniqueTrustedMachineUsers mismatch (-want +got):\n%s", diff)
			}

			// Check unique untrusted machine users
			if diff := cmp.Diff(tt.expectedUniqueUntrustedMachineUsers, tt.config.UniqueUntrustedMachineUsers); diff != "" {
				t.Errorf("UniqueUntrustedMachineUsers mismatch (-want +got):\n%s", diff)
			}

			// Check check name
			if tt.config.CheckName != tt.expectedCheckName {
				t.Errorf("CheckName = %v, want %v", tt.config.CheckName, tt.expectedCheckName)
			}

			// Verify that templates are populated
			if len(tt.config.Templates) == 0 {
				t.Error("Templates should be populated with default templates")
			}

			// Verify that built templates are created
			if len(tt.config.BuiltTemplates) == 0 {
				t.Error("BuiltTemplates should be populated")
			}

			// Verify required templates exist
			requiredTemplates := []string{"footer", "settings", "two_approvals", "no_approval", "require_two_approvals"}
			for _, template := range requiredTemplates {
				if _, exists := tt.config.Templates[template]; !exists {
					t.Errorf("Required template %s not found", template)
				}
			}

			// Verify built templates exist for main template keys
			builtTemplateKeys := []string{"no_approval", "require_two_approvals", "two_approvals"}
			for _, key := range builtTemplateKeys {
				if _, exists := tt.config.BuiltTemplates[key]; !exists {
					t.Errorf("Built template %s not found", key)
				}
			}
		})
	}
}

func TestConfig_Init_TemplateParseError(t *testing.T) {
	t.Parallel()
	config := &Config{
		Templates: map[string]string{
			"no_approval": "{{invalid template syntax}}{{end",
			"footer":      "footer content",
			"settings":    "settings content",
		},
	}

	err := config.Init()
	if err == nil {
		t.Error("Config.Init() should return error for invalid template syntax")
	}
}

func TestConfig_Init_NilTemplates(t *testing.T) {
	t.Parallel()
	config := &Config{
		TrustedApps: []string{"app1[bot]"},
		Templates:   nil,
	}

	err := config.Init()
	if err != nil {
		t.Errorf("Config.Init() with nil Templates should not error, got: %v", err)
	}

	// Verify that templates map is created and populated
	if config.Templates == nil {
		t.Error("Templates map should be created")
	}

	if len(config.Templates) == 0 {
		t.Error("Templates should be populated with default templates")
	}
}
