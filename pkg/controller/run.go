package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
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
	trust := c.input.Config.Trust
	if repo != nil {
		trust = repo.Trust
	}

	// Run validation
	result := c.validate(ctx, logger, ev, trust)
	result.RequestID = req.RequestID

	if err := c.gh.CreateCheckRun(ctx, c.newCheckRunInput(logger, ev, result, trust)); err != nil {
		slogerr.WithError(logger, err).Error("create final check run")
	}
	return nil
}
