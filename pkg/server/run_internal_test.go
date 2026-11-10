package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/controller"
)

type blockingController struct {
	release chan struct{}
	reqs    chan *controller.Request
}

func (c *blockingController) Run(_ context.Context, _ *slog.Logger, req *controller.Request) error {
	<-c.release
	c.reqs <- req
	return nil
}

type countingController struct {
	calls atomic.Int32
}

func (c *countingController) Run(context.Context, *slog.Logger, *controller.Request) error {
	c.calls.Add(1)
	return nil
}

type panicController struct{}

func (c *panicController) Run(context.Context, *slog.Logger, *controller.Request) error {
	panic("test panic")
}

func newTestRequest(t *testing.T) *http.Request {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/webhook", strings.NewReader(`{"action":"submitted"}`))
	req.Header.Set("X-Request-ID", "test-request-id")
	req.Header.Set("X-Github-Event", "pull_request_review")
	return req
}

func TestServer_Run(t *testing.T) {
	t.Setenv("REQUEST_ID_HEADER", "X-Request-ID")
	ctrl := &blockingController{
		release: make(chan struct{}),
		reqs:    make(chan *controller.Request, 1),
	}
	srv, err := New(slog.New(slog.DiscardHandler), ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}

	// The response must be returned while the controller is still processing.
	rec := httptest.NewRecorder()
	srv.Run(rec, newTestRequest(t))
	if rec.Code != http.StatusOK {
		t.Fatalf("status code: wanted %d, got %d", http.StatusOK, rec.Code)
	}

	close(ctrl.release)
	srv.wg.Wait()
	got := <-ctrl.reqs
	want := &controller.Request{
		Body: `{"action":"submitted"}`,
		Headers: map[string]string{
			"X-Request-Id":   "test-request-id",
			"X-Github-Event": "pull_request_review",
		},
		RequestID: "test-request-id",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func TestServer_Run_panic(t *testing.T) {
	t.Parallel()
	srv, err := New(slog.New(slog.DiscardHandler), &panicController{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	srv.Run(rec, newTestRequest(t))
	// A panic in the background must not crash the process.
	srv.wg.Wait()
	if rec.Code != http.StatusOK {
		t.Fatalf("status code: wanted %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestServer_Run_closing(t *testing.T) {
	t.Parallel()
	ctrl := &countingController{}
	srv, err := New(slog.New(slog.DiscardHandler), ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	srv.waitInFlight()

	// Requests after shutdown starts must be rejected without calling the controller.
	rec := httptest.NewRecorder()
	srv.Run(rec, newTestRequest(t))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status code: wanted %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
	srv.wg.Wait()
	if n := ctrl.calls.Load(); n != 0 {
		t.Fatalf("the controller was called %d times after shutdown started", n)
	}
}

func TestServer_waitInFlight(t *testing.T) {
	t.Parallel()
	ctrl := &blockingController{
		release: make(chan struct{}),
		reqs:    make(chan *controller.Request, 1),
	}
	srv, err := New(slog.New(slog.DiscardHandler), ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	srv.Run(httptest.NewRecorder(), newTestRequest(t))

	done := make(chan struct{})
	go func() {
		srv.waitInFlight()
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("waitInFlight returned while a request was still being processed")
	case <-time.After(100 * time.Millisecond):
	}

	close(ctrl.release)
	<-ctrl.reqs
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("waitInFlight didn't return after the request was processed")
	}
}
