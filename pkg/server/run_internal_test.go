package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
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
	srv.waitInFlight(t.Context())

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
	synctest.Test(t, func(t *testing.T) {
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
			srv.waitInFlight(t.Context())
			close(done)
		}()

		// Wait until every goroutine in the bubble is blocked:
		// the controller on release, and waitInFlight on wg.Wait.
		synctest.Wait()
		select {
		case <-done:
			t.Fatal("waitInFlight returned while a request was still being processed")
		default:
		}

		close(ctrl.release)
		<-ctrl.reqs
		synctest.Wait()
		select {
		case <-done:
		default:
			t.Fatal("waitInFlight didn't return after the request was processed")
		}
	})
}

func TestServer_waitInFlight_timeout(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		ctrl := &blockingController{
			release: make(chan struct{}),
			reqs:    make(chan *controller.Request, 1),
		}
		srv, err := New(slog.New(slog.DiscardHandler), ctrl, nil)
		if err != nil {
			t.Fatal(err)
		}
		srv.Run(httptest.NewRecorder(), newTestRequest(t))

		// The time in the bubble is fake, so this doesn't actually take shutdownTimeout.
		ctx, cancel := context.WithTimeout(t.Context(), shutdownTimeout)
		defer cancel()
		start := time.Now()
		srv.waitInFlight(ctx)
		if elapsed := time.Since(start); elapsed != shutdownTimeout {
			t.Fatalf("waitInFlight should give up after %s, but returned after %s", shutdownTimeout, elapsed)
		}

		// Finish the blocked request so that the bubble can exit.
		close(ctrl.release)
		<-ctrl.reqs
		synctest.Wait()
	})
}

func TestServer_Run_tooLargeBody(t *testing.T) {
	t.Parallel()
	ctrl := &countingController{}
	srv, err := New(slog.New(slog.DiscardHandler), ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Generate the body on the fly instead of allocating a large string.
	body := io.LimitReader(&countingReader{}, maxBodySize+1)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/webhook", body)
	rec := httptest.NewRecorder()
	srv.Run(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status code: wanted %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}
	// The request must be unregistered from the WaitGroup, otherwise this blocks forever.
	srv.wg.Wait()
	if n := ctrl.calls.Load(); n != 0 {
		t.Fatalf("the controller was called %d times for a too large body", n)
	}
}

func TestServer_Run_tooLargeContentLength(t *testing.T) {
	t.Parallel()
	ctrl := &countingController{}
	srv, err := New(slog.New(slog.DiscardHandler), ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	// The body must not be read when Content-Length exceeds the limit.
	body := &countingReader{}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/webhook", body)
	req.ContentLength = maxBodySize + 1
	rec := httptest.NewRecorder()
	srv.Run(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status code: wanted %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}
	if body.n != 0 {
		t.Fatalf("the body was read (%d bytes) although Content-Length exceeds the limit", body.n)
	}
	srv.wg.Wait()
	if n := ctrl.calls.Load(); n != 0 {
		t.Fatalf("the controller was called %d times for a too large body", n)
	}
}

// countingReader is an endless reader of zero bytes that counts the bytes read.
type countingReader struct {
	n int
}

func (r *countingReader) Read(p []byte) (int, error) {
	clear(p)
	r.n += len(p)
	return len(p), nil
}

// timeoutReader simulates a read on a connection whose read deadline (ReadTimeout) has passed.
type timeoutReader struct{}

func (timeoutReader) Read([]byte) (int, error) {
	return 0, os.ErrDeadlineExceeded
}

func TestServer_Run_readTimeout(t *testing.T) {
	t.Parallel()
	ctrl := &countingController{}
	srv, err := New(slog.New(slog.DiscardHandler), ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/webhook", timeoutReader{})
	rec := httptest.NewRecorder()
	srv.Run(rec, req)
	if rec.Code != http.StatusRequestTimeout {
		t.Fatalf("status code: wanted %d, got %d", http.StatusRequestTimeout, rec.Code)
	}
	srv.wg.Wait()
	if n := ctrl.calls.Load(); n != 0 {
		t.Fatalf("the controller was called %d times although reading the body timed out", n)
	}
}
