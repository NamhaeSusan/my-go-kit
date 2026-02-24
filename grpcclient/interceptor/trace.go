package interceptor

import (
	"context"
	"strings"

	kitlog "github.com/NamhaeSusan/my-go-kit/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	traceMetadataKey = strings.ToLower(kitlog.TraceHeader)
	spanMetadataKey  = strings.ToLower(kitlog.SpanHeader)
	pSpanMetadataKey = strings.ToLower(kitlog.PSpanHeader)
)

func UnaryClientTraceInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = injectOutgoingTraceMetadata(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamClientTraceInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = injectOutgoingTraceMetadata(ctx)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func injectOutgoingTraceMetadata(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	traceID := kitlog.GetTraceID(ctx)
	if traceID == kitlog.Unknown {
		traceID = kitlog.NewTraceID()
	}

	pSpanID := kitlog.GetSpanID(ctx)
	spanID := kitlog.NewSpanID()

	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		md = md.Copy()
	} else {
		md = metadata.New(nil)
	}

	md.Set(traceMetadataKey, traceID)
	md.Set(spanMetadataKey, spanID)
	md.Set(pSpanMetadataKey, pSpanID)

	return metadata.NewOutgoingContext(ctx, md)
}

func UnaryServerTraceInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = injectIncomingTraceContext(ctx)
		return handler(ctx, req)
	}
}

func StreamServerTraceInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := injectIncomingTraceContext(ss.Context())
		return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: ctx})
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func injectIncomingTraceContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	md, _ := metadata.FromIncomingContext(ctx)

	traceID := firstMetadataValue(md.Get(traceMetadataKey))
	if traceID == "" {
		traceID = kitlog.NewTraceID()
	}

	spanID := firstMetadataValue(md.Get(spanMetadataKey))
	if spanID == "" {
		spanID = kitlog.NewSpanID()
	}

	pSpanID := firstMetadataValue(md.Get(pSpanMetadataKey))
	if pSpanID == "" {
		pSpanID = kitlog.Unknown
	}

	ctx = kitlog.WithTraceID(ctx, traceID)
	ctx = kitlog.WithSpanID(ctx, spanID)
	ctx = kitlog.WithPSpanID(ctx, pSpanID)
	return ctx
}

func firstMetadataValue(values []string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
