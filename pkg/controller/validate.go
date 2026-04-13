package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/validation"
)

func (c *Controller) validate(ctx context.Context, logger *slog.Logger, ev *Event, trust *config.Trust, insecure *config.Insecure) *validation.Result {
	pr, err := c.gh.GetPR(ctx, ev.RepoOwner, ev.RepoName, ev.PRNumber)
	if err != nil {
		return &validation.Result{Error: fmt.Errorf("get a pull request: %w", err).Error()}
	}
	logger.Info("fetched a pull request", "pull_request", pr)

	c.checkMergeCommits(ctx, logger, ev, pr)

	input := &validation.Input{
		PR: pr,
		Trust: &validation.Trust{
			TrustedApps:           trust.UniqueTrustedApps,
			UntrustedMachineUsers: trust.UntrustedMachineUsers,
		},
	}
	if insecure != nil {
		input.Insecure = &validation.Insecure{
			AllowUnsignedCommits:       insecure.AllowUnsignedCommits != nil && *insecure.AllowUnsignedCommits,
			UnsignedCommitApps:         toSet(insecure.UnsignedCommitApps),
			UnsignedCommitMachineUsers: toSet(insecure.UnsignedCommitMachineUsers),
		}
	}
	return c.validator.Run(logger, input)
}

func toSet(s []string) map[string]struct{} {
	m := make(map[string]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
