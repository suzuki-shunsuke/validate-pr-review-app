//nolint:funlen
package controller

import (
	"html/template"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v75/github"
	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/validation"
)

func TestController_newCheckRunInput(t *testing.T) {
	t.Parallel()

	// Setup templates
	templates := map[string]*template.Template{
		"approved":              template.Must(template.New("approved").Parse("PR Approved by {{.Approvers}}")),
		"no_approval":           template.Must(template.New("no_approval").Parse("No approval found")),
		"require_two_approvals": template.Must(template.New("require_two_approvals").Parse("Two approvals required")),
		"error":                 template.Must(template.New("error").Parse("Error: {{.Error}}")),
	}

	tests := []struct {
		name     string
		config   *config.Config
		event    *github.PullRequestReviewEvent
		result   *validation.Result
		expected githubv4.CreateCheckRunInput
	}{
		{
			name: "approved state",
			config: &config.Config{
				CheckName: "test-check",
				Trust: &config.Trust{
					TrustedApps:           []string{"dependabot[bot]"},
					TrustedMachineUsers:   []string{"trusted-user"},
					UntrustedMachineUsers: []string{"untrusted-*"},
				},
				BuiltTemplates: templates,
			},
			event: &github.PullRequestReviewEvent{
				Repo: &github.Repository{
					NodeID: github.String("repo-node-id"),
				},
				PullRequest: &github.PullRequest{
					Head: &github.PullRequestBranch{
						SHA: github.String("abc123"),
					},
				},
			},
			result: &validation.Result{
				State:     validation.StateApproved,
				Approvers: []string{"user1", "user2"},
			},
			expected: githubv4.CreateCheckRunInput{
				RepositoryID: githubv4.String("repo-node-id"),
				HeadSha:      githubv4.GitObjectID("abc123"),
				Name:         githubv4.String("test-check"),
				Status:       &[]githubv4.RequestableCheckStatusState{githubv4.RequestableCheckStatusStateCompleted}[0],
				Conclusion:   &[]githubv4.CheckConclusionState{githubv4.CheckConclusionStateSuccess}[0],
				Output: &githubv4.CheckRunOutput{
					Title:   githubv4.String("Approved"),
					Summary: githubv4.String("PR Approved by [user1 user2]"),
				},
			},
		},
		{
			name: "approval required state",
			config: &config.Config{
				CheckName: "test-check",
				Trust: &config.Trust{
					TrustedApps:           []string{"dependabot[bot]"},
					TrustedMachineUsers:   []string{"trusted-user"},
					UntrustedMachineUsers: []string{"untrusted-*"},
				},
				BuiltTemplates: templates,
			},
			event: &github.PullRequestReviewEvent{
				Repo: &github.Repository{
					NodeID: github.String("repo-node-id"),
				},
				PullRequest: &github.PullRequest{
					Head: &github.PullRequestBranch{
						SHA: github.String("abc123"),
					},
				},
			},
			result: &validation.Result{
				State: validation.StateApprovalIsRequired,
			},
			expected: githubv4.CreateCheckRunInput{
				RepositoryID: githubv4.String("repo-node-id"),
				HeadSha:      githubv4.GitObjectID("abc123"),
				Name:         githubv4.String("test-check"),
				Status:       &[]githubv4.RequestableCheckStatusState{githubv4.RequestableCheckStatusStateCompleted}[0],
				Conclusion:   &[]githubv4.CheckConclusionState{githubv4.CheckConclusionStateFailure}[0],
				Output: &githubv4.CheckRunOutput{
					Title:   githubv4.String("Approvals are required"),
					Summary: githubv4.String("No approval found"),
				},
			},
		},
		{
			name: "two approvals required state",
			config: &config.Config{
				CheckName: "test-check",
				Trust: &config.Trust{
					TrustedApps:           []string{"dependabot[bot]"},
					TrustedMachineUsers:   []string{"trusted-user"},
					UntrustedMachineUsers: []string{"untrusted-*"},
				},
				BuiltTemplates: templates,
			},
			event: &github.PullRequestReviewEvent{
				Repo: &github.Repository{
					NodeID: github.String("repo-node-id"),
				},
				PullRequest: &github.PullRequest{
					Head: &github.PullRequestBranch{
						SHA: github.String("abc123"),
					},
				},
			},
			result: &validation.Result{
				State: validation.StateTwoApprovalsAreRequired,
			},
			expected: githubv4.CreateCheckRunInput{
				RepositoryID: githubv4.String("repo-node-id"),
				HeadSha:      githubv4.GitObjectID("abc123"),
				Name:         githubv4.String("test-check"),
				Status:       &[]githubv4.RequestableCheckStatusState{githubv4.RequestableCheckStatusStateCompleted}[0],
				Conclusion:   &[]githubv4.CheckConclusionState{githubv4.CheckConclusionStateFailure}[0],
				Output: &githubv4.CheckRunOutput{
					Title:   githubv4.String("Two approvals are required"),
					Summary: githubv4.String("Two approvals required"),
				},
			},
		},
		{
			name: "error state",
			config: &config.Config{
				CheckName: "test-check",
				Trust: &config.Trust{
					TrustedApps:           []string{"dependabot[bot]"},
					TrustedMachineUsers:   []string{"trusted-user"},
					UntrustedMachineUsers: []string{"untrusted-*"},
				},
				BuiltTemplates: templates,
			},
			event: &github.PullRequestReviewEvent{
				Repo: &github.Repository{
					NodeID: github.String("repo-node-id"),
				},
				PullRequest: &github.PullRequest{
					Head: &github.PullRequestBranch{
						SHA: github.String("abc123"),
					},
				},
			},
			result: &validation.Result{
				State: validation.StateApproved,
				Error: "test error message",
			},
			expected: githubv4.CreateCheckRunInput{
				RepositoryID: githubv4.String("repo-node-id"),
				HeadSha:      githubv4.GitObjectID("abc123"),
				Name:         githubv4.String("test-check"),
				Status:       &[]githubv4.RequestableCheckStatusState{githubv4.RequestableCheckStatusStateCompleted}[0],
				Conclusion:   &[]githubv4.CheckConclusionState{githubv4.CheckConclusionStateFailure}[0],
				Output: &githubv4.CheckRunOutput{
					Title:   githubv4.String("Internal Error"),
					Summary: githubv4.String("Error: test error message"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			controller := &Controller{
				input: &InputNew{
					Config:  tt.config,
					Version: "test-version",
				},
			}

			logger := slog.Default()
			result := controller.newCheckRunInput(logger, tt.event, tt.result)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("newCheckRunInput() mismatch (-want +got):\n%s", diff)
			}

			// Verify that result fields are populated correctly
			if tt.result.TrustedApps == nil {
				t.Error("TrustedApps should be populated")
			}
			if tt.result.TrustedMachineUsers == nil {
				t.Error("TrustedMachineUsers should be populated")
			}
			if tt.result.UntrustedMachineUsers == nil {
				t.Error("UntrustedMachineUsers should be populated")
			}
			if tt.result.Version == "" {
				t.Error("Version should be populated")
			}
		})
	}
}

func Test_summarize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		result    *validation.Result
		templates map[string]*template.Template
		expected  string
		wantErr   bool
	}{
		{
			name: "approved state",
			result: &validation.Result{
				State:     validation.StateApproved,
				Approvers: []string{"user1", "user2"},
			},
			templates: map[string]*template.Template{
				"approved": template.Must(template.New("approved").Parse("Approved by: {{range .Approvers}}{{.}} {{end}}")),
			},
			expected: "Approved by: user1 user2 ",
			wantErr:  false,
		},
		{
			name: "no approval state",
			result: &validation.Result{
				State: validation.StateApprovalIsRequired,
			},
			templates: map[string]*template.Template{
				"no_approval": template.Must(template.New("no_approval").Parse("No approvals found")),
			},
			expected: "No approvals found",
			wantErr:  false,
		},
		{
			name: "error state",
			result: &validation.Result{
				State: validation.StateApproved,
				Error: "something went wrong",
			},
			templates: map[string]*template.Template{
				"error": template.Must(template.New("error").Parse("Error occurred: {{.Error}}")),
			},
			expected: "Error occurred: something went wrong",
			wantErr:  false,
		},
		{
			name: "missing template",
			result: &validation.Result{
				State: validation.StateApproved,
			},
			templates: map[string]*template.Template{},
			expected:  "",
			wantErr:   true,
		},
		{
			name: "template execution error",
			result: &validation.Result{
				State: validation.StateApproved,
			},
			templates: map[string]*template.Template{
				"approved": template.Must(template.New("approved").Parse("{{.NonExistentField}}")),
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := summarize(tt.result, tt.templates)

			if (err != nil) != tt.wantErr {
				t.Errorf("summarize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.expected {
				t.Errorf("summarize() = %v, want %v", result, tt.expected)
			}
		})
	}
}
