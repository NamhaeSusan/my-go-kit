package log

import (
	"context"
	"strings"
)

const (
	traceFieldName  = "traceId"
	unknownTraceID  = "unknown"
)

type traceIDKeyType struct{}

var traceIDKey traceIDKeyType

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, traceIDKey, traceID)
}

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return unknownTraceID
	}

	traceID, ok := ctx.Value(traceIDKey).(string)
	if !ok || strings.TrimSpace(traceID) == "" {
		return unknownTraceID
	}

	return traceID
}
