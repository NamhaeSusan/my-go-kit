package log

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	initOnce sync.Once
	initErr  error

	baseMu sync.RWMutex
	base   *zap.Logger
)

func Init() error {
	initOnce.Do(func() {
		cfg := zap.NewProductionConfig()
		cfg.Encoding = "json"
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

		logger, err := cfg.Build()
		if err != nil {
			initErr = err
			return
		}

		baseMu.Lock()
		base = logger
		baseMu.Unlock()
	})

	return initErr
}

func Sync() error {
	logger := ensureLogger()
	return logger.Sync()
}

func FromContext(ctx context.Context) *zap.Logger {
	return ensureLogger().With(zap.String(traceFieldName, GetTraceID(ctx)))
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Debug(msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Info(msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Warn(msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Error(msg, fields...)
}

func ensureLogger() *zap.Logger {
	baseMu.RLock()
	logger := base
	baseMu.RUnlock()
	if logger != nil {
		return logger
	}

	if err := Init(); err != nil {
		return zap.NewNop()
	}

	baseMu.RLock()
	logger = base
	baseMu.RUnlock()
	if logger == nil {
		return zap.NewNop()
	}

	return logger
}
