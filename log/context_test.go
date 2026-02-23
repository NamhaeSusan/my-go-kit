package log

import (
	"context"
	"testing"
)

func TestWithTraceIDAndGetTraceID(t *testing.T) {
	ctx := WithTraceID(context.Background(), "abc-123")
	if got := GetTraceID(ctx); got != "abc-123" {
		t.Fatalf("unexpected traceId: %q", got)
	}
}

func TestGetTraceIDMissingReturnsUnknown(t *testing.T) {
	if got := GetTraceID(context.Background()); got != Unknown {
		t.Fatalf("unexpected traceId: %q", got)
	}
}

func TestGetTraceIDNilContextReturnsUnknown(t *testing.T) {
	if got := GetTraceID(nil); got != Unknown { //nolint:staticcheck // intentional nil context test
		t.Fatalf("unexpected traceId: %q", got)
	}
}

func TestGetTraceIDEmptyReturnsUnknown(t *testing.T) {
	ctx := WithTraceID(context.Background(), "")
	if got := GetTraceID(ctx); got != Unknown {
		t.Fatalf("unexpected traceId: %q", got)
	}
}

func TestWithSpanIDAndGetSpanID(t *testing.T) {
	ctx := WithSpanID(context.Background(), "span-1")
	if got := GetSpanID(ctx); got != "span-1" {
		t.Fatalf("unexpected spanId: %q", got)
	}
}

func TestGetSpanIDMissingReturnsUnknown(t *testing.T) {
	if got := GetSpanID(context.Background()); got != Unknown {
		t.Fatalf("unexpected spanId: %q", got)
	}
}

func TestWithPSpanIDAndGetPSpanID(t *testing.T) {
	ctx := WithPSpanID(context.Background(), "pspan-1")
	if got := GetPSpanID(ctx); got != "pspan-1" {
		t.Fatalf("unexpected pSpanId: %q", got)
	}
}

func TestGetPSpanIDMissingReturnsUnknown(t *testing.T) {
	if got := GetPSpanID(context.Background()); got != Unknown {
		t.Fatalf("unexpected pSpanId: %q", got)
	}
}
