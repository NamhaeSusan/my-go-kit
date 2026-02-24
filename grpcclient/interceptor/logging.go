package interceptor

import (
	"context"
	"strings"
	"time"

	kitlog "github.com/NamhaeSusan/my-go-kit/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	logTypeFieldName = "log_type"
	logTypeGRPC      = "grpc"
)

func UnaryClientLoggingInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		fullMethod string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		err := invoker(ctx, fullMethod, req, reply, cc, opts...)
		logGRPCCall(ctx, fullMethod, time.Since(start), err)
		return err
	}
}

func StreamClientLoggingInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		fullMethod string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		start := time.Now()
		stream, err := streamer(ctx, desc, cc, fullMethod, opts...)
		logGRPCCall(ctx, fullMethod, time.Since(start), err)
		return stream, err
	}
}

func UnaryServerLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		fullMethod := ""
		if info != nil {
			fullMethod = info.FullMethod
		}
		logGRPCCall(ctx, fullMethod, time.Since(start), err)
		return resp, err
	}
}

func StreamServerLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		err := handler(srv, ss)
		fullMethod := ""
		if info != nil {
			fullMethod = info.FullMethod
		}
		logGRPCCall(ss.Context(), fullMethod, time.Since(start), err)
		return err
	}
}

func logGRPCCall(ctx context.Context, fullMethod string, elapsed time.Duration, err error) {
	service, method := splitGRPCMethod(fullMethod)
	fields := append(
		kitlog.FromContext(ctx),
		zap.Int64("elapsed", elapsed.Milliseconds()),
		zap.String("method", method),
		zap.String("service", service),
		zap.String("grpc_code", status.Code(err).String()),
		zap.String(logTypeFieldName, logTypeGRPC),
	)

	zap.L().Info("grpc request", fields...)
}

func splitGRPCMethod(fullMethod string) (string, string) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(fullMethod), "/")
	if trimmed == "" {
		return "unknown", "unknown"
	}

	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "unknown", "unknown"
	}
	return parts[0], parts[1]
}
