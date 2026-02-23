package log

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	initOnce sync.Once
	initErr  error
)

func Init(logFilePath string) error {
	initOnce.Do(func() {
		atomicLevel := zap.NewAtomicLevelAt(zapcore.InfoLevel)

		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			FunctionKey:    zapcore.OmitKey,
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		var cores []zapcore.Core

		consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), atomicLevel)
		cores = append(cores, consoleCore)

		if logFilePath != "" {
			if err := os.MkdirAll(filepath.Dir(logFilePath), os.ModePerm); err != nil {
				initErr = err
				return
			}
			fileLogger := &lumberjack.Logger{
				Filename:  logFilePath,
				MaxSize:   1024, // 1MB
				MaxAge:    7,    // 7 days
				Compress:  true,
				LocalTime: true,
			}

			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGHUP)
			go func() {
				for range c {
					if err := fileLogger.Rotate(); err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "log rotate failed: %v\n", err)
					}
				}
			}()

			fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
			fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileLogger), atomicLevel)
			cores = append(cores, fileCore)
		}

		core := zapcore.NewTee(cores...)
		logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
		zap.ReplaceGlobals(logger)
	})

	return initErr
}

func Sync() error {
	return zap.L().Sync()
}

func FromContext(ctx context.Context) []zap.Field {
	return []zap.Field{
		zap.String(traceFieldName, GetTraceID(ctx)),
		zap.String(spanIDFieldName, GetSpanID(ctx)),
		zap.String(pSpanIDFieldName, GetPSpanID(ctx)),
	}
}

func Debugf(ctx context.Context, msgFormat string, args ...any) {
	zap.L().Debug(formatMessage(msgFormat, args...), FromContext(ctx)...)
}

func Infof(ctx context.Context, msgFormat string, args ...any) {
	zap.L().Info(formatMessage(msgFormat, args...), FromContext(ctx)...)
}

func Warnf(ctx context.Context, msgFormat string, args ...any) {
	zap.L().Warn(formatMessage(msgFormat, args...), FromContext(ctx)...)
}

func Errorf(ctx context.Context, msgFormat string, args ...any) {
	zap.L().Error(formatMessage(msgFormat, args...), FromContext(ctx)...)
}

func formatMessage(msgFormat string, args ...any) string {
	if len(args) == 0 {
		return msgFormat
	}
	return fmt.Sprintf(msgFormat, args...)
}
