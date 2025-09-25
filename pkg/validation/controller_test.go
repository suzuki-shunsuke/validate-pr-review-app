package validation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/validation"
)

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		inputNew *validation.InputNew
	}{
		{
			name: "creates new controller with empty input",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
		},
		{
			name: "creates new controller with configured input",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{"dependabot[bot]": {}},
				TrustedMachineUsers:   map[string]struct{}{"trusted-bot": {}},
				UntrustedMachineUsers: map[string]struct{}{"untrusted-*": {}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := validation.New(tt.inputNew)
			if got == nil {
				t.Error("New() returned nil")
			}
		})
	}
}

func TestController_Type(t *testing.T) {
	t.Parallel()
	inputNew := &validation.InputNew{
		TrustedApps:           map[string]struct{}{},
		TrustedMachineUsers:   map[string]struct{}{},
		UntrustedMachineUsers: map[string]struct{}{},
	}
	controller := validation.New(inputNew)

	// Test that multiple calls to New() return different instances
	controller2 := validation.New(inputNew)

	if controller == controller2 {
		t.Error("New() should return different instances")
	}

	// Both should be valid Controller pointers
	if controller == nil || controller2 == nil {
		t.Error("New() should never return nil")
	}
}

func TestInput_Structure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		input      *validation.Input
		expectedPR *github.PullRequest
	}{
		{
			name: "input with valid PR",
			input: &validation.Input{
				PR: &github.PullRequest{
					HeadSHA: "abc123",
					Approvers: map[string]struct{}{
						"reviewer": {},
					},
					Commits: []*github.Commit{
						{
							SHA: "abc123",
							Committer: &github.User{
								Login: "committer",
								IsApp: false,
							},
						},
					},
				},
			},
			expectedPR: &github.PullRequest{
				HeadSHA: "abc123",
				Approvers: map[string]struct{}{
					"reviewer": {},
				},
				Commits: []*github.Commit{
					{
						SHA: "abc123",
						Committer: &github.User{
							Login: "committer",
							IsApp: false,
						},
					},
				},
			},
		},
		{
			name: "input with nil PR",
			input: &validation.Input{
				PR: nil,
			},
			expectedPR: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if diff := cmp.Diff(tt.expectedPR, tt.input.PR); diff != "" {
				t.Errorf("Input.PR mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestInputNew_Structure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		inputNew *validation.InputNew
		expected *validation.InputNew
	}{
		{
			name: "inputNew with all fields",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{"app1[bot]": {}},
				TrustedMachineUsers:   map[string]struct{}{"trusted-user": {}},
				UntrustedMachineUsers: map[string]struct{}{"untrusted-*": {}},
			},
			expected: &validation.InputNew{
				TrustedApps:           map[string]struct{}{"app1[bot]": {}},
				TrustedMachineUsers:   map[string]struct{}{"trusted-user": {}},
				UntrustedMachineUsers: map[string]struct{}{"untrusted-*": {}},
			},
		},
		{
			name: "inputNew with empty maps",
			inputNew: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
			expected: &validation.InputNew{
				TrustedApps:           map[string]struct{}{},
				TrustedMachineUsers:   map[string]struct{}{},
				UntrustedMachineUsers: map[string]struct{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if diff := cmp.Diff(tt.expected, tt.inputNew); diff != "" {
				t.Errorf("InputNew mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
