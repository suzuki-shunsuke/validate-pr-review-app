package server

import (
	"context"
	"errors"
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
	// readTimeout bounds reading the whole request including the body.
	// A request is registered as in-flight before its body is read, and the server waits for in-flight requests on shutdown without a deadline,
	// so a client that stalls while sending the body would otherwise block shutdown forever.
	readTimeout = 30 * time.Second
	defaultPort = "8080"
	maxBodySize = 25 << 20 // 25 MB
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
		ReadTimeout:       readTimeout,
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
	h.waitInFlight()
	h.logger.Info("server exited")
}

// waitInFlight stops accepting new webhook requests and waits for in-flight ones,
// including ones processed in the background.
// They aren't canceled because the Check Run should still be created to avoid leaving PRs in an indeterminate state.
// srv.Shutdown may return before all handlers finish (e.g. on timeout), so handlers can still call acquire concurrently.
func (h *Server) waitInFlight() {
	h.mu.Lock()
	h.closing = true
	h.mu.Unlock()
	h.logger.Info("waiting for in-flight webhook events to be processed")
	h.wg.Wait()
}

// acquire registers an in-flight webhook request.
// It returns false if the server is shutting down.
// The caller must call h.wg.Done() when the request is processed if it returns true.
func (h *Server) acquire() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closing {
		return false
	}
	h.wg.Add(1)
	return true
}

func ready(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}

func (h *Server) Run(w http.ResponseWriter, r *http.Request) {
	logger, requestID := h.newLogger(r)
	// Register the request before reading the body so that the server waits for it on shutdown
	// even if srv.Shutdown times out while the body is being read.
	if !h.acquire() {
		// srv.Shutdown closes the listener first, so new connections are refused during shutdown.
		// This path is reached only by requests on connections accepted before that,
		// whose handlers start after srv.Shutdown has returned (e.g. on timeout).
		// Return 503 so that GitHub treats the delivery as failed.
		// GitHub doesn't redeliver failed deliveries automatically, but they can be redelivered manually from Recent Deliveries.
		logger.Warn("reject the request because the server is shutting down")
		http.Error(w, "the server is shutting down", http.StatusServiceUnavailable)
		return
	}
	background := false
	defer func() {
		if !background {
			h.wg.Done()
		}
	}()

	// Limit the body size because the body is held in memory until it is processed in the background.
	// GitHub caps webhook payloads at 25 MB.
	// https://docs.github.com/en/webhooks/webhook-events-and-payloads#payload-cap
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodySize))
	if err != nil {
		if maxBytesErr, ok := errors.AsType[*http.MaxBytesError](err); ok {
			logger.Warn("the request body is too large", "limit", maxBytesErr.Limit)
			http.Error(w, "the request body is too large", http.StatusRequestEntityTooLarge)
			return
		}
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
	// The background goroutine takes over the h.wg.Done() call.
	background = true
	go func() { //nolint:contextcheck
		defer h.wg.Done()
		h.handle(logger, req)
	}()
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
