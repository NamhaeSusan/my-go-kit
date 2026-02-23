package httpclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	kitlog "github.com/NamhaeSusan/my-go-kit/log"
)

func TestClientInjectsTraceFromContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(kitlog.TraceHeader); got != "trace-ctx-1" {
			t.Fatalf("unexpected trace header: %q", got)
		}
		// spanID는 새로 생성되어야 한다 (원본 span-ctx-1이 아닌 값).
		if got := r.Header.Get(kitlog.SpanHeader); got == "" || got == "span-ctx-1" {
			t.Fatalf("span header should be newly generated, got: %q", got)
		}
		// 현재 spanID가 다음 서비스의 pSpanID로 전파된다.
		if got := r.Header.Get(kitlog.PSpanHeader); got != "span-ctx-1" {
			t.Fatalf("unexpected pspan header: %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(Config{HTTPClient: server.Client()})
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	ctx := kitlog.WithTraceID(context.Background(), "trace-ctx-1")
	ctx = kitlog.WithSpanID(ctx, "span-ctx-1")
	ctx = kitlog.WithPSpanID(ctx, "pspan-ctx-1")

	resp, err := client.Do(ctx, req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	_ = resp.Body.Close()
}

func TestClientGeneratesTraceWhenMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(kitlog.TraceHeader); got == "" {
			t.Fatalf("unexpected trace header: %q", got)
		}
		if got := r.Header.Get(kitlog.SpanHeader); got == "" {
			t.Fatalf("unexpected span header: %q", got)
		}
		if got := r.Header.Get(kitlog.PSpanHeader); got != "unknown" {
			t.Fatalf("unexpected pspan header: %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(Config{
		HTTPClient: server.Client(),
	})

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	_ = resp.Body.Close()
}

func TestClientRetriesOnRetryableStatus(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(Config{
		HTTPClient: server.Client(),
		Retry: RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   1 * time.Millisecond,
			MaxDelay:    2 * time.Millisecond,
		},
	})

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	_ = resp.Body.Close()

	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Fatalf("unexpected attempt count: %d", got)
	}
}

func TestClientDoesNotRetryNonRetryableMethodByDefault(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := New(Config{
		HTTPClient: server.Client(),
		Retry: RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   1 * time.Millisecond,
			MaxDelay:    2 * time.Millisecond,
		},
	})

	req, err := http.NewRequest(http.MethodPost, server.URL, strings.NewReader("payload"))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	_ = resp.Body.Close()

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("unexpected attempt count: %d", got)
	}
}

func TestClientReturnsErrorOnNilRequest(t *testing.T) {
	client := New(Config{})
	if _, err := client.Do(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestClientStopsOnContextCanceled(t *testing.T) {
	rt := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, context.Canceled
	})
	httpClient := &http.Client{Transport: rt}
	client := New(Config{
		HTTPClient: httpClient,
		Retry: RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   1 * time.Millisecond,
			MaxDelay:    2 * time.Millisecond,
		},
	})

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	if _, err = client.Do(context.Background(), req); !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %v", err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
