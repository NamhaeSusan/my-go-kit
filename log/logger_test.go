package log

import (
	"context"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestFromContextAlwaysAddsTraceID(t *testing.T) {
	resetGlobalLoggerState()

	core, logs := observer.New(zapcore.DebugLevel)
	setBaseLoggerForTest(zap.New(core))

	FromContext(context.Background()).Info("without trace")
	FromContext(WithTraceID(context.Background(), "t-1")).Info("with trace")

	entries := logs.All()
	if len(entries) != 2 {
		t.Fatalf("unexpected log count: %d", len(entries))
	}

	if got := entries[0].ContextMap()[traceFieldName]; got != unknownTraceID {
		t.Fatalf("unexpected traceId: %#v", got)
	}

	if got := entries[1].ContextMap()[traceFieldName]; got != "t-1" {
		t.Fatalf("unexpected traceId: %#v", got)
	}
}

func TestLazyInitBeforeInitCall(t *testing.T) {
	resetGlobalLoggerState()

	Info(context.Background(), "lazy init")

	baseMu.RLock()
	logger := base
	baseMu.RUnlock()
	if logger == nil {
		t.Fatal("base logger should be initialized")
	}
}

func TestLevelHelpersDoNotPanic(t *testing.T) {
	resetGlobalLoggerState()

	core, logs := observer.New(zapcore.DebugLevel)
	setBaseLoggerForTest(zap.New(core))

	Debug(context.Background(), "debug")
	Info(WithTraceID(context.Background(), "abc"), "info")
	Warn(context.Background(), "warn")
	Error(context.Background(), "error")

	if got := logs.Len(); got != 4 {
		t.Fatalf("unexpected log count: %d", got)
	}
}

func resetGlobalLoggerState() {
	baseMu.Lock()
	base = nil
	baseMu.Unlock()
	initErr = nil
	initOnce = sync.Once{}
}

func setBaseLoggerForTest(logger *zap.Logger) {
	baseMu.Lock()
	base = logger
	baseMu.Unlock()
}
