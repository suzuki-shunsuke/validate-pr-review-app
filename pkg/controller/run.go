package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
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
	result := c.validate(ctx, logger, ev, &trust, &insecure)
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
		if repo.TrustedMachineUsers != nil {
			trust.TrustedMachineUsers = repo.TrustedMachineUsers
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
	if repo != nil {
		insecure = *repo
	}
	return insecure
}
