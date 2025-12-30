//nolint:funlen
package config_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/validation"
)

func TestTemplates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		result   *validation.Result
		template string
		wantErr  bool
		wantText string
	}{
		{
			name: "two approvals",
			result: &validation.Result{
				State:     validation.StateApproved,
				Approvers: []string{"user1", "user2"},
				Version:   "v0.0.1",
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

[This check is created by Validate PR Review App](https://github.com/suzuki-shunsuke/validate-pr-review-app).

- Version: v0.0.1
- Request ID: unknown
`,
		},
		{
			name: "one approval",
			result: &validation.Result{
				State:                 validation.StateApproved,
				Approvers:             []string{"user1"},
				TrustedApps:           []string{"dependabot[bot]", "renovate[bot]"},
				TrustedMachineUsers:   []string{"foo-bot"},
				UntrustedMachineUsers: []string{"*-bot"},
				RequestID:             "req-12345",
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

[This check is created by Validate PR Review App](https://github.com/suzuki-shunsuke/validate-pr-review-app).

- Version: unknown
- Request ID: req-12345
`,
		},
		{
			name: "error",
			result: &validation.Result{
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

[This check is created by Validate PR Review App](https://github.com/suzuki-shunsuke/validate-pr-review-app).

- Version: unknown
- Request ID: unknown
`,
		},
		{
			name: "no approval",
			result: &validation.Result{
				State: validation.StateApprovalIsRequired,
			},
			template: "no_approval",
			wantText: `This commit has no approvals.
Approvals are required.

## Settings

Trusted Apps: Nothing

Untrusted Machine Users: Nothing

Trusted Machine Users: Nothing

---

[This check is created by Validate PR Review App](https://github.com/suzuki-shunsuke/validate-pr-review-app).

- Version: unknown
- Request ID: unknown
`,
		},
		{
			name: "require two approvals",
			result: &validation.Result{
				State:        validation.StateTwoApprovalsAreRequired,
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

` + "`foo` approved this pull request, but it's a self-approval. `foo` pushes commits to this pull request." + `

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

[This check is created by Validate PR Review App](https://github.com/suzuki-shunsuke/validate-pr-review-app).

- Version: unknown
- Request ID: unknown
`,
		},
	}
	cfg := &config.Config{
		AWS: &config.AWS{
			SecretID: "validate-pr-review-app",
		},
	}
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
