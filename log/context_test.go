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
	if got := GetTraceID(context.Background()); got != unknownTraceID {
		t.Fatalf("unexpected traceId: %q", got)
	}
}

func TestGetTraceIDNilContextReturnsUnknown(t *testing.T) {
	if got := GetTraceID(nil); got != unknownTraceID {
		t.Fatalf("unexpected traceId: %q", got)
	}
}

func TestGetTraceIDEmptyReturnsUnknown(t *testing.T) {
	ctx := WithTraceID(context.Background(), "")
	if got := GetTraceID(ctx); got != unknownTraceID {
		t.Fatalf("unexpected traceId: %q", got)
	}
}
