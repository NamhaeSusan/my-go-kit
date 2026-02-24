package interceptor

import (
	"context"
	"testing"

	kitlog "github.com/NamhaeSusan/my-go-kit/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryClientLoggingInterceptor_WritesExpectedFields(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	prev := zap.L()
	zap.ReplaceGlobals(logger)
	t.Cleanup(func() { zap.ReplaceGlobals(prev) })

	interceptor := UnaryClientLoggingInterceptor()
	ctx := kitlog.WithTraceID(context.Background(), "trace-1")
	ctx = kitlog.WithSpanID(ctx, "span-1")
	ctx = kitlog.WithPSpanID(ctx, "pspan-1")

	err := interceptor(ctx, "/sample.EchoService/Ping", nil, nil, nil, func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		return status.Error(codes.NotFound, "not found")
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("unexpected error code: %v", status.Code(err))
	}

	if logs.Len() != 1 {
		t.Fatalf("expected exactly one log entry, got: %d", logs.Len())
	}

	entry := logs.All()[0]
	fields := entry.ContextMap()
	if got := fields["method"]; got != "Ping" {
		t.Fatalf("unexpected method: %#v", got)
	}
	if got := fields["service"]; got != "sample.EchoService" {
		t.Fatalf("unexpected service: %#v", got)
	}
	if got := fields["grpc_code"]; got != "NotFound" {
		t.Fatalf("unexpected grpc_code: %#v", got)
	}
	if got := fields[logTypeFieldName]; got != logTypeGRPC {
		t.Fatalf("unexpected log_type: %#v", got)
	}
	if got := fields["elapsed"]; got == nil {
		t.Fatal("elapsed field should exist")
	}
}

func TestStreamClientLoggingInterceptor_WritesOKCode(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	prev := zap.L()
	zap.ReplaceGlobals(logger)
	t.Cleanup(func() { zap.ReplaceGlobals(prev) })

	interceptor := StreamClientLoggingInterceptor()
	_, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/sample.Chat/Join", func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if logs.Len() != 1 {
		t.Fatalf("expected exactly one log entry, got: %d", logs.Len())
	}

	fields := logs.All()[0].ContextMap()
	if got := fields["method"]; got != "Join" {
		t.Fatalf("unexpected method: %#v", got)
	}
	if got := fields["service"]; got != "sample.Chat" {
		t.Fatalf("unexpected service: %#v", got)
	}
	if got := fields["grpc_code"]; got != "OK" {
		t.Fatalf("unexpected grpc_code: %#v", got)
	}
	if got := fields[logTypeFieldName]; got != logTypeGRPC {
		t.Fatalf("unexpected log_type: %#v", got)
	}
}

func TestSplitGRPCMethod(t *testing.T) {
	service, method := splitGRPCMethod("/pkg.Service/Call")
	if service != "pkg.Service" || method != "Call" {
		t.Fatalf("unexpected split result: service=%q method=%q", service, method)
	}

	service, method = splitGRPCMethod("invalid")
	if service != "unknown" || method != "unknown" {
		t.Fatalf("unexpected fallback split result: service=%q method=%q", service, method)
	}
}

func TestUnaryServerLoggingInterceptor_WritesExpectedFields(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	prev := zap.L()
	zap.ReplaceGlobals(logger)
	t.Cleanup(func() { zap.ReplaceGlobals(prev) })

	interceptor := UnaryServerLoggingInterceptor()
	ctx := kitlog.WithTraceID(context.Background(), "srv-trace")
	ctx = kitlog.WithSpanID(ctx, "srv-span")
	ctx = kitlog.WithPSpanID(ctx, "srv-pspan")

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "/sample.Server/Handle",
	}, func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.PermissionDenied, "denied")
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unexpected error code: %v", status.Code(err))
	}
	if logs.Len() != 1 {
		t.Fatalf("expected exactly one log entry, got: %d", logs.Len())
	}

	fields := logs.All()[0].ContextMap()
	if got := fields["method"]; got != "Handle" {
		t.Fatalf("unexpected method: %#v", got)
	}
	if got := fields["service"]; got != "sample.Server" {
		t.Fatalf("unexpected service: %#v", got)
	}
	if got := fields["grpc_code"]; got != "PermissionDenied" {
		t.Fatalf("unexpected grpc_code: %#v", got)
	}
	if got := fields[logTypeFieldName]; got != logTypeGRPC {
		t.Fatalf("unexpected log_type: %#v", got)
	}
}

func TestStreamServerLoggingInterceptor_WritesOKCode(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	prev := zap.L()
	zap.ReplaceGlobals(logger)
	t.Cleanup(func() { zap.ReplaceGlobals(prev) })

	interceptor := StreamServerLoggingInterceptor()
	err := interceptor(nil, &fakeServerStreamForLogging{ctx: context.Background()}, &grpc.StreamServerInfo{
		FullMethod: "/sample.Stream/Chat",
	}, func(srv any, stream grpc.ServerStream) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logs.Len() != 1 {
		t.Fatalf("expected exactly one log entry, got: %d", logs.Len())
	}

	fields := logs.All()[0].ContextMap()
	if got := fields["method"]; got != "Chat" {
		t.Fatalf("unexpected method: %#v", got)
	}
	if got := fields["service"]; got != "sample.Stream" {
		t.Fatalf("unexpected service: %#v", got)
	}
	if got := fields["grpc_code"]; got != "OK" {
		t.Fatalf("unexpected grpc_code: %#v", got)
	}
	if got := fields[logTypeFieldName]; got != logTypeGRPC {
		t.Fatalf("unexpected log_type: %#v", got)
	}
}

type fakeServerStreamForLogging struct {
	grpc.ServerStream
	ctx context.Context
}

func (f *fakeServerStreamForLogging) Context() context.Context {
	return f.ctx
}
