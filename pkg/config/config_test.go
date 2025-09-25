//nolint:funlen
package config_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

func TestConfig_Init(t *testing.T) { //nolint:gocognit,cyclop
	t.Parallel()
	tests := []struct {
		name                                string
		config                              *config.Config
		expectedUniqueTrustedApps           map[string]struct{}
		expectedUniqueTrustedMachineUsers   map[string]struct{}
		expectedUniqueUntrustedMachineUsers map[string]struct{}
		expectedCheckName                   string
		wantErr                             bool
	}{
		{
			name: "basic initialization",
			config: &config.Config{
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
			config: &config.Config{
				Templates: map[string]string{},
			},
			expectedUniqueTrustedApps: map[string]struct{}{
				"dependabot[bot]": {},
				"renovate[bot]":   {},
			},
			expectedUniqueTrustedMachineUsers:   map[string]struct{}{},
			expectedUniqueUntrustedMachineUsers: map[string]struct{}{},
			expectedCheckName:                   "verify-approval", // default value
		},
		{
			name: "duplicate entries in arrays",
			config: &config.Config{
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
			expectedCheckName: "verify-approval",
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
			requiredTemplates := []string{"footer", "settings", "approved", "no_approval", "require_two_approvals"}
			for _, template := range requiredTemplates {
				if _, exists := tt.config.Templates[template]; !exists {
					t.Errorf("Required template %s not found", template)
				}
			}

			// Verify built templates exist for main template keys
			builtTemplateKeys := []string{"no_approval", "require_two_approvals", "approved"}
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
	config := &config.Config{
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
	config := &config.Config{
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

func TestTemplates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		result   *config.Result
		template string
		wantErr  bool
		wantText string
	}{
		{
			name: "two approvals",
			result: &config.Result{
				State:     config.StateApproved,
				Approvers: []string{"user1", "user2"},
			},
			template: "approved",
			wantErr:  false,
			wantText: `The pull request has been approved.

Approvers:

- user1
- user2

## Settings

Trusted Apps: Nothing

Untrusted Machine Users: Nothing

Trusted Machine Users: Nothing

---

[This check is created by Enforce PR Review App](https://github.com/suzuki-shunsuke/enforce-pr-review-app).
`,
		},
		{
			name: "one approval",
			result: &config.Result{
				State:                 config.StateApproved,
				Approvers:             []string{"user1"},
				TrustedApps:           []string{"dependabot[bot]", "renovate[bot]"},
				TrustedMachineUsers:   []string{"foo-bot"},
				UntrustedMachineUsers: []string{"*-bot"},
			},
			template: "approved",
			wantErr:  false,
			wantText: `The pull request has been approved.

Approvers:

- user1

## Settings

Trusted Apps:

- dependabot[bot]
- renovate[bot]

Untrusted Machine Users:
- *-bot

Trusted Machine Users:

- foo-bot

---

[This check is created by Enforce PR Review App](https://github.com/suzuki-shunsuke/enforce-pr-review-app).
`,
		},
		{
			name: "error",
			result: &config.Result{
				Error:                 "failed to fetch pr",
				TrustedApps:           []string{"dependabot[bot]", "renovate[bot]"},
				TrustedMachineUsers:   []string{"foo-bot"},
				UntrustedMachineUsers: []string{"*-bot"},
			},
			template: "error",
			wantErr:  false,
			wantText: `failed to fetch pr

## Settings

Trusted Apps:

- dependabot[bot]
- renovate[bot]

Untrusted Machine Users:
- *-bot

Trusted Machine Users:

- foo-bot

---

[This check is created by Enforce PR Review App](https://github.com/suzuki-shunsuke/enforce-pr-review-app).
`,
		},
		{
			name: "no approval",
			result: &config.Result{
				State: config.StateApprovalIsRequired,
			},
			template: "no_approval",
			wantText: `This commit has no approvals.
Approvals are required.

## Settings

Trusted Apps: Nothing

Untrusted Machine Users: Nothing

Trusted Machine Users: Nothing

---

[This check is created by Enforce PR Review App](https://github.com/suzuki-shunsuke/enforce-pr-review-app).
`,
		},
		{
			name: "require two approvals",
			result: &config.Result{
				State:        config.StateTwoApprovalsAreRequired,
				SelfApprover: "foo",
				Approvers:    []string{"user1"},
				IgnoredApprovers: []*github.IgnoredApproval{
					{
						Login:                  "foo-bot",
						IsUntrustedMachineUser: true,
					},
				},
				UntrustedMachineUsers: []string{"*-bot"},
				UntrustedCommits: []*github.UntrustedCommit{
					{
						Login:                  "foo-bot",
						SHA:                    "xxx",
						IsUntrustedMachineUser: true,
					},
				},
			},
			template: "require_two_approvals",
			wantText: `This pull request requires two approvals because:

` + "`foo`" + ` approved this pull request, but it's a self-approval.

The following commits are untrusted, so two approvals are required.

- xxx foo-bot The committer is an untrusted machine user.

## :warning: Some approvals are ignored

Approvals from GitHub Apps and Untrusted Machine Users are ignored.

Approvals from the following approvers are ignored:
- foo-bot Untrusted Machine User

## Settings

Trusted Apps: Nothing
Untrusted Machine Users:
- *-bot

Trusted Machine Users: Nothing

---

[This check is created by Enforce PR Review App](https://github.com/suzuki-shunsuke/enforce-pr-review-app).
`,
		},
	}
	cfg := &config.Config{}
	if err := cfg.Init(); err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tpl, ok := cfg.BuiltTemplates[tt.template]
			if !ok {
				t.Fatalf("template %s not found", tt.template)
			}
			buf := &bytes.Buffer{}
			if err := tpl.Execute(buf, tt.result); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.wantText, buf.String()); diff != "" {
				t.Errorf("text mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
