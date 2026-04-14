package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/validation"
)

func (c *Controller) Run(ctx context.Context, logger *slog.Logger, req *Request) error {
	logger.Debug("Starting a request", "request", req)
	// Validate the request
	ev := c.verifyWebhook(logger, req)
	if ev == nil {
		return nil
	}
	logger = logger.With(
		"repository", ev.RepoFullName,
		"pr_number", ev.PRNumber,
		"sha", ev.HeadSHA,
		"pr_url", fmt.Sprintf("https://github.com/%s/pull/%d", ev.RepoFullName, ev.PRNumber),
	)

	if ignore(logger, ev) {
		return nil
	}
	repo := c.input.Config.GetRepo(ev.RepoFullName)
	if repo != nil && repo.Ignored {
		logger.Info("ignore the event because the repository is ignored in the config", "repository", ev.RepoFullName)
		return nil
	}
	var repoTrust *config.Trust
	var repoInsecure *config.Insecure
	if repo != nil {
		repoTrust = repo.Trust
		repoInsecure = repo.Insecure
	}
	trust := mergeTrust(c.input.Config.Trust, repoTrust)
	insecure := mergeInsecure(c.input.Config.Insecure, repoInsecure)
	trust.Init()

	// Run validation
	var result *validation.Result
	if ev.EventType == eventPullRequest {
		result = c.carryForwardCheck(ctx, logger, ev, &trust, &insecure)
		if result == nil {
			logger.Info("carry-forward check not applicable, skipping")
			return nil
		}
	} else {
		result = c.validate(ctx, logger, ev, &trust, &insecure)
	}
	result.RequestID = req.RequestID

	if err := c.gh.CreateCheckRun(ctx, c.newCheckRunInput(logger, ev, result, &trust, &insecure)); err != nil {
		slogerr.WithError(logger, err).Error("create final check run")
	}
	return nil
}

func mergeTrust(global *config.Trust, repo *config.Trust) config.Trust {
	var trust config.Trust
	if global != nil {
		trust = *global
	}
	if repo != nil {
		if repo.TrustedApps != nil {
			trust.TrustedApps = repo.TrustedApps
		}
		if repo.UntrustedMachineUsers != nil {
			trust.UntrustedMachineUsers = repo.UntrustedMachineUsers
		}
	}
	return trust
}

func mergeInsecure(global *config.Insecure, repo *config.Insecure) config.Insecure {
	var insecure config.Insecure
	if global != nil {
		insecure = *global
	}
	if repo == nil {
		return insecure
	}
	if repo.UnsignedCommitApps != nil {
		insecure.UnsignedCommitApps = repo.UnsignedCommitApps
		insecure.AllowUnsignedCommits = new(false)
	}
	if repo.UnsignedCommitMachineUsers != nil {
		insecure.UnsignedCommitMachineUsers = repo.UnsignedCommitMachineUsers
		insecure.AllowUnsignedCommits = new(false)
	}
	if repo.AllowUnsignedCommits != nil {
		insecure.AllowUnsignedCommits = repo.AllowUnsignedCommits
		if *repo.AllowUnsignedCommits {
			// If repo.AllowUnsignedCommits is true, it overrides global UnsignedCommitApps and UnsignedCommitMachineUsers.
			insecure.UnsignedCommitApps = nil
			insecure.UnsignedCommitMachineUsers = nil
		}
	}
	return insecure
}
