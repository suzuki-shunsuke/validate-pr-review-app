package aws

import (
	"context"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
)

func (h *Handler) do(ctx context.Context, req *controller.Request) {
	logger := h.newLogger(ctx)
	if err := h.controller.Run(ctx, logger, req); err != nil {
		slogerr.WithError(logger, err).Error("handle request")
	}
}
