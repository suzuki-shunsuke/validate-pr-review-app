package validation

import (
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

type Input struct {
	Config *config.Config
	PR     *github.PullRequest
}
