package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
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
	// Wait for webhook events being processed in the background.
	// Don't cancel them because the Check Run should still be created to avoid leaving PRs in an indeterminate state.
	h.logger.Info("waiting for in-flight webhook events to be processed")
	h.wg.Wait()
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

	req := &controller.Request{
		Body:      string(body),
		Headers:   convertHeaders(r.Header),
		RequestID: requestID,
	}

	// Process the webhook event in the background and return 200 immediately regardless of the result.
	// GitHub terminates the connection and treats the delivery as failed if the server doesn't respond within 10 seconds.
	// https://docs.github.com/en/webhooks/using-webhooks/best-practices-for-using-webhooks#respond-within-10-seconds
	// The processing result isn't returned to GitHub anyway, so responding before processing loses nothing.
	// The request context isn't passed because it is canceled when the handler returns.
	h.wg.Go(func() { //nolint:contextcheck
		h.handle(logger, req)
	})
	w.WriteHeader(http.StatusOK)
}

func (h *Server) handle(logger *slog.Logger, req *controller.Request) {
	// net/http recovers panics in handlers, but not in goroutines started by them.
	// Recover here so that a panic doesn't crash the whole server.
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic while handling a request", "panic", r, "stack", string(debug.Stack()))
		}
	}()

	// Use context.Background() instead of r.Context()
	// https://github.com/suzuki-shunsuke/validate-pr-review-app/pull/401
	// The request is processed after the response is returned, so r.Context() is already canceled.
	// Why context.Background() is correct?
	// There are two sources of context cancellation, and neither should abort processing:
	// 1. The HTTP request
	//   The response is returned before processing, and GitHub's connection may be closed at any time.
	//   This does not indicate intent to abort processing.
	// 2. Server shutdown (k8s pod/node rotation)
	//   validate-pr-review-app runs as an HTTP server on Kubernetes, etc.
	//   Pod rollouts, node rotations, and other infrastructure events are unavoidable and will trigger server graceful shutdown, canceling the server's context.
	//   However, these events do not mean we want to abort in-flight operations — the Check Run should still be created to avoid leaving PRs in an indeterminate state.
	//   The server waits for in-flight operations on shutdown.
	//
	// Both cases differ from a CLI tool where Ctrl-C is an intentional user action to stop processing.
	// In the webhook server case, neither the sender (GitHub) nor the infrastructure (k8s) intends to cancel the application's business logic.
	if err := h.controller.Run(context.Background(), logger, req); err != nil {
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
