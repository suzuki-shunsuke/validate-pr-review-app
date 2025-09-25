package validation

import (
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

type Controller struct {
	input *InputNew
}

type InputNew struct {
	TrustedApps           map[string]struct{}
	TrustedMachineUsers   map[string]struct{}
	UntrustedMachineUsers map[string]struct{}
}

func New(input *InputNew) *Controller {
	return &Controller{
		input: input,
	}
}

type Input struct {
	PR *github.PullRequest
}
