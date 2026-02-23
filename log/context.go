package log

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

const (
	traceFieldName   = "traceId"
	spanIDFieldName  = "spanId"
	pSpanIDFieldName = "pSpanId"

	// Unknown is the fallback value when a trace/span ID is not set.
	Unknown = "unknown"

	TraceHeader = "X-Trace-Id"
	SpanHeader  = "X-Span-Id"
	PSpanHeader = "X-PSpan-Id"
)

type traceIDKeyType struct{}
type spanIDKeyType struct{}
type pSpanIDKeyType struct{}

var TraceIDKey traceIDKeyType
var SpanIDKey spanIDKeyType
var PSpanIDKey pSpanIDKeyType

func NewTraceID() string {
	u := uuid.New()
	return hex.EncodeToString(u[:16])
}

func NewSpanID() string {
	u := uuid.New()
	return hex.EncodeToString(u[:8])
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, TraceIDKey, traceID)
}

func WithSpanID(ctx context.Context, spanID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, SpanIDKey, spanID)
}

func WithPSpanID(ctx context.Context, pSpanID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, PSpanIDKey, pSpanID)
}

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return Unknown
	}

	traceID, ok := ctx.Value(TraceIDKey).(string)
	if !ok || strings.TrimSpace(traceID) == "" {
		return Unknown
	}

	return traceID
}

func GetSpanID(ctx context.Context) string {
	if ctx == nil {
		return Unknown
	}

	spanID, ok := ctx.Value(SpanIDKey).(string)
	if !ok || strings.TrimSpace(spanID) == "" {
		return Unknown
	}

	return spanID
}

func GetPSpanID(ctx context.Context) string {
	if ctx == nil {
		return Unknown
	}

	pSpanID, ok := ctx.Value(PSpanIDKey).(string)
	if !ok || strings.TrimSpace(pSpanID) == "" {
		return Unknown
	}

	return pSpanID
}
