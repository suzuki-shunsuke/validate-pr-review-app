package validation

import (
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
)

type Validator struct {
	input *InputNew
}

type InputNew struct {
	// TrustedApps           map[string]struct{}
	// UntrustedMachineUsers []string
}

func New(input *InputNew) *Validator {
	return &Validator{
		input: input,
	}
}

type Input struct {
	PR       *github.PullRequest
	Trust    *Trust
	Insecure *Insecure
}

type Insecure struct {
	AllowUnsignedCommits       bool
	UnsignedCommitApps         map[string]struct{}
	UnsignedCommitMachineUsers map[string]struct{}
}

type Trust struct {
	TrustedApps           map[string]struct{}
	UntrustedMachineUsers []string
}
