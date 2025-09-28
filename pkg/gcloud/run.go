package gcloud

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
)

const (
	// readHeaderTimeout defines the maximum time allowed to read request headers
	readHeaderTimeout = 10 * time.Second
	// defaultPort is the default HTTP server port when PORT environment variable is not set
	defaultPort = "8080"
)

// Start begins the HTTP server for Google Cloud Functions with graceful shutdown support.
// It listens on the PORT environment variable (defaulting to 8080) and handles incoming webhook requests.
// The server shuts down gracefully when the provided context is cancelled.
func (h *Handler) Start(ctx context.Context) {
	// Set up HTTP router with webhook handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Run)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Configure HTTP server with security timeouts
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Start server in a goroutine for non-blocking operation
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slogerr.WithError(h.logger, err).Error("start HTTP server")
		}
	}()

	// Wait for context cancellation signal
	<-ctx.Done()

	h.logger.Info("shutting down gracefully...")

	// Perform graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //nolint:mnd
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck
		slogerr.WithError(h.logger, err).Error("shutdown failed")
	}
	h.logger.Info("server exited")
}

// Run handles incoming HTTP requests from Google Cloud Functions HTTP triggers.
// It reads the request body, extracts headers, and forwards the request to the controller for processing.
// This function serves as the main entry point for GitHub webhook events in Cloud Functions.
func (h *Handler) Run(w http.ResponseWriter, r *http.Request) {
	// Create a logger with request tracing information
	logger, requestID := h.newLogger(r)

	// Read the entire request body for webhook processing
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slogerr.WithError(logger, err).Error("read request body")
		http.Error(w, "failed to read the request body", http.StatusBadRequest)
		return
	}

	// Forward the request to the controller for GitHub webhook processing
	if err := h.controller.Run(r.Context(), logger, &controller.Request{
		Body:      string(body),
		Headers:   convertHeaders(r.Header),
		RequestID: requestID,
	}); err != nil {
		slogerr.WithError(logger, err).Error("handle request")
	}
}

// newLogger creates a context-aware logger that includes the Google Cloud trace ID for request tracing.
// It extracts the trace ID from the X-Cloud-Trace-Context header and adds it to the logger.
func (h *Handler) newLogger(r *http.Request) (*slog.Logger, string) {
	logger := h.logger
	// Extract Google Cloud trace context for request correlation
	if requestID := r.Header.Get("X-Cloud-Trace-Context"); requestID != "" {
		return logger.With("request_id", requestID), requestID
	}
	logger.Warn("X-Cloud-Trace-Context header is not set")
	return logger, ""
}

// convertHeaders converts HTTP headers from http.Header to a map[string]string format.
// This conversion is needed for compatibility with the controller.Request structure.
func convertHeaders(headers http.Header) map[string]string {
	hs := make(map[string]string, len(headers))
	for k := range headers {
		hs[k] = headers.Get(k)
	}
	return hs
}
