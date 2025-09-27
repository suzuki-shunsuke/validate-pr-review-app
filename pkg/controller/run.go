package controller

import (
	"context"
	"log/slog"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (c *Controller) Run(ctx context.Context, logger *slog.Logger, req *Request) error {
	logger.Info("Starting a request", "request", req)
	defer logger.Info("Ending a request")

	// Validate the request
	ev := c.verifyWebhook(logger, req)
	if ev == nil {
		return nil
	}

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

	if err := c.gh.CreateCheckRun(ctx, c.newCheckRunInput(logger, ev, result, trust)); err != nil {
		slogerr.WithError(logger, err).Error("create final check run")
	}
	return nil
}
