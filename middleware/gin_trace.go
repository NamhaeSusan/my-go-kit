package middleware

import (
	"strings"

	kitlog "github.com/NamhaeSusan/my-go-kit/log"
	"github.com/gin-gonic/gin"
)

const (
	TraceIDContextKey = "traceId"
	SpanIDContextKey  = "spanId"
	PSpanIDContextKey = "pSpanId"
)

type TraceIDConfig struct {
	HeaderName        string
	SpanHeaderName    string
	PSpanHeaderName   string
	SetResponseHeader *bool
}

func GinTraceID() gin.HandlerFunc {
	return GinTraceIDWithConfig(TraceIDConfig{
		HeaderName:        kitlog.TraceHeader,
		SpanHeaderName:    kitlog.SpanHeader,
		PSpanHeaderName:   kitlog.PSpanHeader,
		SetResponseHeader: new(true),
	})
}

func GinTraceIDWithConfig(cfg TraceIDConfig) gin.HandlerFunc {
	traceHeader := cfg.HeaderName
	if traceHeader == "" {
		traceHeader = kitlog.TraceHeader
	}

	spanHeader := cfg.SpanHeaderName
	if spanHeader == "" {
		spanHeader = kitlog.SpanHeader
	}

	pSpanHeader := cfg.PSpanHeaderName
	if pSpanHeader == "" {
		pSpanHeader = kitlog.PSpanHeader
	}

	setResponseHeader := true
	if cfg.SetResponseHeader != nil {
		setResponseHeader = *cfg.SetResponseHeader
	}

	return func(c *gin.Context) {
		traceID := strings.TrimSpace(c.GetHeader(traceHeader))
		if traceID == "" {
			traceID = kitlog.NewTraceID()
		}

		spanID := strings.TrimSpace(c.GetHeader(spanHeader))
		if spanID == "" {
			spanID = kitlog.NewSpanID()
		}

		pSpanID := strings.TrimSpace(c.GetHeader(pSpanHeader))
		if pSpanID == "" {
			pSpanID = kitlog.Unknown
		}

		ctx := kitlog.WithTraceID(c.Request.Context(), traceID)
		ctx = kitlog.WithSpanID(ctx, spanID)
		ctx = kitlog.WithPSpanID(ctx, pSpanID)

		c.Request = c.Request.WithContext(ctx)

		c.Set(TraceIDContextKey, traceID)
		c.Set(SpanIDContextKey, spanID)
		c.Set(PSpanIDContextKey, pSpanID)

		if setResponseHeader {
			c.Header(traceHeader, traceID)
			c.Header(spanHeader, spanID)
			c.Header(pSpanHeader, pSpanID)
		}

		c.Next()
	}
}
