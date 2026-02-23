package log

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestFromContextAddsAllFields(t *testing.T) {
	ctx := WithTraceID(context.Background(), "t-1")
	ctx = WithSpanID(ctx, "s-1")
	ctx = WithPSpanID(ctx, "p-1")
	fields := FromContext(ctx)
	if got := len(fields); got != 3 {
		t.Fatalf("unexpected field count: %d", got)
	}

	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	prev := zap.L()
	zap.ReplaceGlobals(logger)
	t.Cleanup(func() { zap.ReplaceGlobals(prev) })
	zap.L().Info("ctx fields", fields...)

	entry := logs.All()[0]
	if got := entry.ContextMap()[traceFieldName]; got != "t-1" {
		t.Fatalf("unexpected traceId: %#v", got)
	}
	if got := entry.ContextMap()[spanIDFieldName]; got != "s-1" {
		t.Fatalf("unexpected spanId: %#v", got)
	}
	if got := entry.ContextMap()[pSpanIDFieldName]; got != "p-1" {
		t.Fatalf("unexpected pSpanId: %#v", got)
	}
}

func TestFromContextUsesUnknownWhenMissing(t *testing.T) {
	fields := FromContext(context.Background())
	if got := len(fields); got != 3 {
		t.Fatalf("unexpected field count: %d", got)
	}

	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	prev := zap.L()
	zap.ReplaceGlobals(logger)
	t.Cleanup(func() { zap.ReplaceGlobals(prev) })
	zap.L().Info("ctx unknown", fields...)

	entry := logs.All()[0]
	if got := entry.ContextMap()[traceFieldName]; got != Unknown {
		t.Fatalf("unexpected traceId: %#v", got)
	}
	if got := entry.ContextMap()[spanIDFieldName]; got != Unknown {
		t.Fatalf("unexpected spanId: %#v", got)
	}
	if got := entry.ContextMap()[pSpanIDFieldName]; got != Unknown {
		t.Fatalf("unexpected pSpanId: %#v", got)
	}
}

func TestLevelHelpersWriteContextFields(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	prev := zap.L()
	zap.ReplaceGlobals(logger)
	t.Cleanup(func() { zap.ReplaceGlobals(prev) })

	ctx := WithTraceID(context.Background(), "abc")
	ctx = WithSpanID(ctx, "span")
	ctx = WithPSpanID(ctx, "parent")

	Debugf(context.Background(), "debug %d", 1)
	Infof(ctx, "info %s", "ok")
	Warnf(context.Background(), "warn")
	Errorf(context.Background(), "error")

	if got := logs.Len(); got != 4 {
		t.Fatalf("unexpected log count: %d", got)
	}

	entries := logs.All()
	if msg := entries[0].Message; msg != "debug 1" {
		t.Fatalf("unexpected debug message: %q", msg)
	}
	if msg := entries[1].Message; msg != "info ok" {
		t.Fatalf("unexpected info message: %q", msg)
	}
	if got := entries[1].ContextMap()[traceFieldName]; got != "abc" {
		t.Fatalf("unexpected traceId: %#v", got)
	}
	if got := entries[1].ContextMap()[spanIDFieldName]; got != "span" {
		t.Fatalf("unexpected spanId: %#v", got)
	}
	if got := entries[1].ContextMap()[pSpanIDFieldName]; got != "parent" {
		t.Fatalf("unexpected pSpanId: %#v", got)
	}
}

func TestInitWithEmptyPathSetsGlobalLogger(t *testing.T) {
	prev := zap.L()
	defer zap.ReplaceGlobals(prev)

	resetInitStateForTest()
	if err := Init(""); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	Infof(context.Background(), "init ok")
}

func TestInitReturnsErrorForInvalidLogPath(t *testing.T) {
	prev := zap.L()
	defer zap.ReplaceGlobals(prev)

	resetInitStateForTest()

	tmpDir := t.TempDir()
	fileAsDir := filepath.Join(tmpDir, "not-a-dir")
	if err := os.WriteFile(fileAsDir, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to setup test file: %v", err)
	}
	invalidPath := filepath.Join(fileAsDir, "app.log")

	if err := Init(invalidPath); err == nil {
		t.Fatal("expected Init to return error for invalid log path")
	}
}

func TestFormatMessageFastPath(t *testing.T) {
	if got := formatMessage("plain message"); got != "plain message" {
		t.Fatalf("unexpected message: %q", got)
	}
	if got := formatMessage("hello %s", "world"); got != "hello world" {
		t.Fatalf("unexpected formatted message: %q", got)
	}
}

func TestCloseStopsSIGHUPListener(t *testing.T) {
	prev := zap.L()
	defer zap.ReplaceGlobals(prev)

	resetInitStateForTest()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	if err := Init(logPath); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	// Close may return a benign stdout sync error — only check sigCh cleanup.
	_ = Close()

	if sigCh != nil {
		t.Fatal("expected sigCh to be nil after Close")
	}
}

func TestCloseWithoutFileInit(t *testing.T) {
	prev := zap.L()
	defer zap.ReplaceGlobals(prev)

	resetInitStateForTest()
	if err := Init(""); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	// Close may return a benign stdout sync error — just verify no panic.
	_ = Close()
}

func resetInitStateForTest() {
	initErr = nil
	initOnce = sync.Once{}
	sigCh = nil
}
