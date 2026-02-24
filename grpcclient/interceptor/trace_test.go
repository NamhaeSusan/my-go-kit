package interceptor

import (
	"context"
	"encoding/hex"
	"testing"

	kitlog "github.com/NamhaeSusan/my-go-kit/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryClientTraceInterceptor_InjectsOutgoingMetadata(t *testing.T) {
	interceptor := UnaryClientTraceInterceptor()

	ctx := kitlog.WithTraceID(context.Background(), "trace-ctx-1")
	ctx = kitlog.WithSpanID(ctx, "span-ctx-1")
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("x-existing", "value"))

	var gotCtx context.Context
	err := interceptor(ctx, "/svc/method", nil, nil, nil, func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		gotCtx = ctx
		return nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	md, ok := metadata.FromOutgoingContext(gotCtx)
	if !ok {
		t.Fatal("outgoing metadata should exist")
	}
	if got := md.Get("x-existing"); len(got) != 1 || got[0] != "value" {
		t.Fatalf("existing metadata should be preserved, got: %v", got)
	}
	if got := firstMetadataValue(md.Get(traceMetadataKey)); got != "trace-ctx-1" {
		t.Fatalf("unexpected trace id: %q", got)
	}
	if got := firstMetadataValue(md.Get(pSpanMetadataKey)); got != "span-ctx-1" {
		t.Fatalf("unexpected pspan id: %q", got)
	}
	if got := firstMetadataValue(md.Get(spanMetadataKey)); got == "" || got == "span-ctx-1" {
		t.Fatalf("span id should be regenerated, got: %q", got)
	}
}

func TestUnaryClientTraceInterceptor_GeneratesWhenMissing(t *testing.T) {
	interceptor := UnaryClientTraceInterceptor()

	var gotCtx context.Context
	err := interceptor(nil, "/svc/method", nil, nil, nil, func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		gotCtx = ctx
		return nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	md, ok := metadata.FromOutgoingContext(gotCtx)
	if !ok {
		t.Fatal("outgoing metadata should exist")
	}
	traceID := firstMetadataValue(md.Get(traceMetadataKey))
	spanID := firstMetadataValue(md.Get(spanMetadataKey))
	pSpanID := firstMetadataValue(md.Get(pSpanMetadataKey))
	if !isHexLen(traceID, 32) {
		t.Fatalf("trace id should be generated hex32, got: %q", traceID)
	}
	if !isHexLen(spanID, 16) {
		t.Fatalf("span id should be generated hex16, got: %q", spanID)
	}
	if pSpanID != kitlog.Unknown {
		t.Fatalf("pspan should fallback to unknown, got: %q", pSpanID)
	}
}

func TestStreamClientTraceInterceptor_InjectsOutgoingMetadata(t *testing.T) {
	interceptor := StreamClientTraceInterceptor()

	ctx := kitlog.WithTraceID(context.Background(), "trace-ctx-2")
	ctx = kitlog.WithSpanID(ctx, "span-ctx-2")

	var gotCtx context.Context
	_, err := interceptor(ctx, &grpc.StreamDesc{}, nil, "/svc/stream", func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		gotCtx = ctx
		return nil, nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	md, ok := metadata.FromOutgoingContext(gotCtx)
	if !ok {
		t.Fatal("outgoing metadata should exist")
	}
	if got := firstMetadataValue(md.Get(traceMetadataKey)); got != "trace-ctx-2" {
		t.Fatalf("unexpected trace id: %q", got)
	}
	if got := firstMetadataValue(md.Get(pSpanMetadataKey)); got != "span-ctx-2" {
		t.Fatalf("unexpected pspan id: %q", got)
	}
}

func TestUnaryServerTraceInterceptor_InjectsIncomingContext(t *testing.T) {
	interceptor := UnaryServerTraceInterceptor()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		traceMetadataKey, "incoming-trace",
		spanMetadataKey, "incoming-span",
		pSpanMetadataKey, "incoming-pspan",
	))

	var gotCtx context.Context
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		gotCtx = ctx
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	if got := kitlog.GetTraceID(gotCtx); got != "incoming-trace" {
		t.Fatalf("unexpected trace id: %q", got)
	}
	if got := kitlog.GetSpanID(gotCtx); got != "incoming-span" {
		t.Fatalf("unexpected span id: %q", got)
	}
	if got := kitlog.GetPSpanID(gotCtx); got != "incoming-pspan" {
		t.Fatalf("unexpected pspan id: %q", got)
	}
}

func TestStreamServerTraceInterceptor_GeneratesFallbackValues(t *testing.T) {
	interceptor := StreamServerTraceInterceptor()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		traceMetadataKey, "   ",
		spanMetadataKey, "",
	))

	ss := &fakeServerStream{ctx: ctx}
	err := interceptor(nil, ss, &grpc.StreamServerInfo{}, func(srv any, stream grpc.ServerStream) error {
		injected := stream.Context()
		traceID := kitlog.GetTraceID(injected)
		spanID := kitlog.GetSpanID(injected)
		pSpanID := kitlog.GetPSpanID(injected)

		if !isHexLen(traceID, 32) {
			t.Fatalf("trace id should be generated hex32, got: %q", traceID)
		}
		if !isHexLen(spanID, 16) {
			t.Fatalf("span id should be generated hex16, got: %q", spanID)
		}
		if pSpanID != kitlog.Unknown {
			t.Fatalf("pspan should fallback to unknown, got: %q", pSpanID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
}

func TestFirstMetadataValue_TrimsAndSkipsEmpty(t *testing.T) {
	got := firstMetadataValue([]string{"", "  ", " value ", "ignored"})
	if got != "value" {
		t.Fatalf("unexpected metadata value: %q", got)
	}

	if got := firstMetadataValue([]string{"", "   "}); got != "" {
		t.Fatalf("expected empty value, got: %q", got)
	}
}

type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (f *fakeServerStream) Context() context.Context {
	return f.ctx
}

func isHexLen(s string, l int) bool {
	if len(s) != l {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
