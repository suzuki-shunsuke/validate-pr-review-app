package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
)

const (
	readHeaderTimeout = 10 * time.Second
	defaultPort       = "8080"
)

func (h *Server) Start(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", h.Run)
	mux.HandleFunc("/ready", ready)
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slogerr.WithError(h.logger, err).Error("start HTTP server")
		}
	}()

	<-ctx.Done()

	h.logger.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //nolint:mnd
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck
		slogerr.WithError(h.logger, err).Error("shutdown failed")
	}
	h.logger.Info("server exited")
}

func ready(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}

func (h *Server) Run(w http.ResponseWriter, r *http.Request) {
	logger, requestID := h.newLogger(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slogerr.WithError(logger, err).Error("read request body")
		http.Error(w, "failed to read the request body", http.StatusBadRequest)
		return
	}

	// Use context.Background() instead of r.Context()
	// https://github.com/suzuki-shunsuke/validate-pr-review-app/pull/401
	// When GitHub closes the webhook HTTP connection (after a few seconds timeout), r.Context() is canceled.
	// So using r.Context() may cancel Create Check Run.
	// Why context.Background() is correct?
	// There are two sources of context cancellation, and neither should abort processing:
	// 1. GitHub webhook connection timeout
	//   GitHub may close the HTTP connection after a few seconds, but this does not indicate intent to abort processing — it's simply an HTTP connection timeout.
	//   The application should continue processing the webhook event and create the check run regardless.
	// 2. Server shutdown (k8s pod/node rotation)
	//   validate-pr-review-app runs as an HTTP server on AWS Lambda, Kubernetes, etc.
	//   Pod rollouts, node rotations, and other infrastructure events are unavoidable and will trigger server graceful shutdown, canceling the server's context.
	//   However, these events do not mean we want to abort in-flight operations — the Check Run should still be created to avoid leaving PRs in an indeterminate state.
	//
	// Both cases differ from a CLI tool where Ctrl-C is an intentional user action to stop processing.
	// In the webhook server case, neither the sender (GitHub) nor the infrastructure (k8s) intends to cancel the application's business logic.
	if err := h.controller.Run(context.Background(), logger, &controller.Request{ //nolint:contextcheck
		Body:      string(body),
		Headers:   convertHeaders(r.Header),
		RequestID: requestID,
	}); err != nil {
		slogerr.WithError(logger, err).Error("handle request")
	}
}

func (h *Server) getRequestID(r *http.Request) string {
	header := os.Getenv("REQUEST_ID_HEADER")
	if header == "" {
		header = "X-Cloud-Trace-Context"
	}
	if requestID := r.Header.Get(header); requestID != "" {
		return requestID
	}
	return uuid.New().String()
}

func (h *Server) newLogger(r *http.Request) (*slog.Logger, string) {
	logger := h.logger
	if requestID := h.getRequestID(r); requestID != "" {
		return logger.With("request_id", requestID), requestID
	}
	logger.Debug("the request header is not set")
	return logger, ""
}

func convertHeaders(headers http.Header) map[string]string {
	hs := make(map[string]string, len(headers))
	for k := range headers {
		hs[k] = headers.Get(k)
	}
	return hs
}
