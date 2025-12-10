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
	mux.HandleFunc("/", h.Run)
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

func (h *Server) Run(w http.ResponseWriter, r *http.Request) {
	logger, requestID := h.newLogger(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slogerr.WithError(logger, err).Error("read request body")
		http.Error(w, "failed to read the request body", http.StatusBadRequest)
		return
	}
	if err := h.controller.Run(r.Context(), logger, &controller.Request{
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
