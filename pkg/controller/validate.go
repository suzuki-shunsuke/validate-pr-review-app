package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/go-github/v75/github"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/validation"
)

func (c *Controller) validate(ctx context.Context, logger *slog.Logger, ev *github.PullRequestReviewEvent, trust *config.Trust) *validation.Result {
	repo := ev.GetRepo()
	owner := repo.GetOwner().GetLogin()
	pr, err := c.gh.GetPR(ctx, owner, repo.GetName(), ev.GetPullRequest().GetNumber())
	if err != nil {
		return &validation.Result{Error: fmt.Errorf("get a pull request: %w", err).Error()}
	}
	logger.Info("fetched a pull request", "pull_request", pr)
	return c.validator.Run(logger, &validation.Input{
		PR: pr,
		Trust: &validation.Trust{
			TrustedApps:           trust.UniqueTrustedApps,
			TrustedMachineUsers:   trust.UniqueTrustedMachineUsers,
			UntrustedMachineUsers: trust.UniqueUntrustedMachineUsers,
		},
	})
}
