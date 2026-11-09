package server

import (
	"context"
	"log/slog"
	"sync"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
)

type Server struct {
	logger     *slog.Logger
	config     *config.Config
	controller Controller
	// wg tracks webhook events being processed in the background
	// so that the server waits for them on shutdown.
	wg sync.WaitGroup
}

type Controller interface {
	Run(ctx context.Context, logger *slog.Logger, req *controller.Request) error
}

func New(logger *slog.Logger, ctrl Controller, cfg *config.Config) (*Server, error) {
	return &Server{
		logger:     logger,
		config:     cfg,
		controller: ctrl,
	}, nil
}
