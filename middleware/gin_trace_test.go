package middleware

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	kitlog "github.com/NamhaeSusan/my-go-kit/log"
	"github.com/gin-gonic/gin"
)

func TestGinTraceID_UsesIncomingHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := buildTestRouter(GinTraceID())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(kitlog.TraceHeader, "incoming-trace")
	req.Header.Set(kitlog.SpanHeader, "incoming-span")
	req.Header.Set(kitlog.PSpanHeader, "incoming-pspan")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	body := decodeBody(t, rec)
	if body.CtxTrace != "incoming-trace" || body.GinTrace != "incoming-trace" {
		t.Fatalf("unexpected trace values: ctx=%q gin=%q", body.CtxTrace, body.GinTrace)
	}
	if body.CtxSpan != "incoming-span" || body.GinSpan != "incoming-span" {
		t.Fatalf("unexpected span values: ctx=%q gin=%q", body.CtxSpan, body.GinSpan)
	}
	if body.CtxPSpan != "incoming-pspan" || body.GinPSpan != "incoming-pspan" {
		t.Fatalf("unexpected pspan values: ctx=%q gin=%q", body.CtxPSpan, body.GinPSpan)
	}

	if got := rec.Header().Get(kitlog.TraceHeader); got != "incoming-trace" {
		t.Fatalf("unexpected trace response header: %q", got)
	}
	if got := rec.Header().Get(kitlog.SpanHeader); got != "incoming-span" {
		t.Fatalf("unexpected span response header: %q", got)
	}
	if got := rec.Header().Get(kitlog.PSpanHeader); got != "incoming-pspan" {
		t.Fatalf("unexpected pspan response header: %q", got)
	}
}

func TestGinTraceID_GeneratesTraceSpanAndUnknownPSpanWhenMissingOrBlank(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := buildTestRouter(GinTraceID())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(kitlog.TraceHeader, " ")
	req.Header.Set(kitlog.SpanHeader, "")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	body := decodeBody(t, rec)
	if !isHex32(body.CtxTrace) || !isHex16(body.CtxSpan) {
		t.Fatalf("expected generated values, got: %+v", body)
	}
	if body.CtxPSpan != kitlog.Unknown {
		t.Fatalf("expected unknown pspan, got: %q", body.CtxPSpan)
	}
	if body.CtxTrace != body.GinTrace || body.CtxSpan != body.GinSpan || body.CtxPSpan != body.GinPSpan {
		t.Fatalf("ctx/gin mismatch: %+v", body)
	}
	if rec.Header().Get(kitlog.TraceHeader) != body.CtxTrace {
		t.Fatalf("trace header mismatch: %q", rec.Header().Get(kitlog.TraceHeader))
	}
	if rec.Header().Get(kitlog.SpanHeader) != body.CtxSpan {
		t.Fatalf("span header mismatch: %q", rec.Header().Get(kitlog.SpanHeader))
	}
	if rec.Header().Get(kitlog.PSpanHeader) != kitlog.Unknown {
		t.Fatalf("pspan header mismatch: %q", rec.Header().Get(kitlog.PSpanHeader))
	}
}

func TestGinTraceIDWithConfig_CustomHeadersAndNoResponseHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := buildTestRouter(GinTraceIDWithConfig(TraceIDConfig{
		HeaderName:        "X-Request-Id",
		SpanHeaderName:    "X-Span",
		PSpanHeaderName:   "X-Parent-Span",
		SetResponseHeader: new(false),
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "req-trace")
	req.Header.Set("X-Span", "req-span")
	req.Header.Set("X-Parent-Span", "req-pspan")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	body := decodeBody(t, rec)
	if body.CtxTrace != "req-trace" || body.CtxSpan != "req-span" || body.CtxPSpan != "req-pspan" {
		t.Fatalf("unexpected context values: %+v", body)
	}

	if got := rec.Header().Get("X-Request-Id"); got != "" {
		t.Fatalf("expected empty response trace header, got: %q", got)
	}
	if got := rec.Header().Get("X-Span"); got != "" {
		t.Fatalf("expected empty response span header, got: %q", got)
	}
	if got := rec.Header().Get("X-Parent-Span"); got != "" {
		t.Fatalf("expected empty response pspan header, got: %q", got)
	}
}

func TestGinTraceIDWithConfig_DefaultsWhenHeaderNamesEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := buildTestRouter(GinTraceIDWithConfig(TraceIDConfig{}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(kitlog.TraceHeader, "default-trace")
	req.Header.Set(kitlog.SpanHeader, "default-span")
	req.Header.Set(kitlog.PSpanHeader, "default-pspan")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	body := decodeBody(t, rec)
	if body.CtxTrace != "default-trace" || body.CtxSpan != "default-span" || body.CtxPSpan != "default-pspan" {
		t.Fatalf("unexpected defaults behavior: %+v", body)
	}
}

type responseBody struct {
	CtxTrace string `json:"ctxTrace"`
	GinTrace string `json:"ginTrace"`
	CtxSpan  string `json:"ctxSpan"`
	GinSpan  string `json:"ginSpan"`
	CtxPSpan string `json:"ctxPSpan"`
	GinPSpan string `json:"ginPSpan"`
}

func buildTestRouter(mw gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(mw)
	router.GET("/", func(c *gin.Context) {
		ginTrace, _ := c.Get(TraceIDContextKey)
		ginSpan, _ := c.Get(SpanIDContextKey)
		ginPSpan, _ := c.Get(PSpanIDContextKey)
		c.JSON(http.StatusOK, responseBody{
			CtxTrace: kitlog.GetTraceID(c.Request.Context()),
			GinTrace: asString(ginTrace),
			CtxSpan:  kitlog.GetSpanID(c.Request.Context()),
			GinSpan:  asString(ginSpan),
			CtxPSpan: kitlog.GetPSpanID(c.Request.Context()),
			GinPSpan: asString(ginPSpan),
		})
	})
	return router
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func isHex32(s string) bool {
	if len(s) != 32 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func isHex16(s string) bool {
	if len(s) != 16 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) responseBody {
	t.Helper()

	var body responseBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return body
}
