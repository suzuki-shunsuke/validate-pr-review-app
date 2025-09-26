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
	ev, err := c.validateRequest(logger, req)
	if err != nil {
		slogerr.WithError(logger, err).Warn("validate request")
		return nil
	}

	if ignore(logger, ev) {
		return nil
	}

	// Run validation
	result := c.validate(ctx, logger, ev)

	if err := c.gh.CreateCheckRun(ctx, c.newCheckRunInput(logger, ev, result)); err != nil {
		slogerr.WithError(logger, err).Error("create final check run")
	}
	return nil
}
