package log

import (
	"context"
	"strings"
)

const (
	traceFieldName   = "traceId"
	spanIDFieldName  = "spanId"
	pSpanIDFieldName = "pSpanId"
	unknown          = "unknown"
)

type traceIDKeyType struct{}
type spanIDKeyType struct{}
type pSpanIDKeyType struct{}

var traceIDKey traceIDKeyType
var spanIDKey spanIDKeyType
var pSpanIDKey pSpanIDKeyType

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, traceIDKey, traceID)
}

func WithSpanID(ctx context.Context, spanID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, spanIDKey, spanID)
}

func WithPSpanID(ctx context.Context, pSpanID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, pSpanIDKey, pSpanID)
}

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return unknown
	}

	traceID, ok := ctx.Value(traceIDKey).(string)
	if !ok || strings.TrimSpace(traceID) == "" {
		return unknown
	}

	return traceID
}

func GetSpanID(ctx context.Context) string {
	if ctx == nil {
		return unknown
	}

	spanID, ok := ctx.Value(spanIDKey).(string)
	if !ok || strings.TrimSpace(spanID) == "" {
		return unknown
	}

	return spanID
}

func GetPSpanID(ctx context.Context) string {
	if ctx == nil {
		return unknown
	}

	pSpanID, ok := ctx.Value(pSpanIDKey).(string)
	if !ok || strings.TrimSpace(pSpanID) == "" {
		return unknown
	}

	return pSpanID
}
