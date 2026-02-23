package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	kitlog "github.com/NamhaeSusan/my-go-kit/log"
)



type Config struct {
	HTTPClient *http.Client
	Retry      RetryConfig
}

type RetryConfig struct {
	MaxAttempts          int
	BaseDelay            time.Duration
	MaxDelay             time.Duration
	RetryableStatuses    []int
	RetryableHTTPMethods []string
}

type Client struct {
	httpClient        *http.Client
	maxAttempts       int
	baseDelay         time.Duration
	maxDelay          time.Duration
	retryableStatuses map[int]struct{}
	retryableMethods  map[string]struct{}
}

func New(cfg Config) *Client {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	maxAttempts := cfg.Retry.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	baseDelay := cfg.Retry.BaseDelay
	if baseDelay <= 0 {
		baseDelay = 100 * time.Millisecond
	}

	maxDelay := cfg.Retry.MaxDelay
	if maxDelay <= 0 {
		maxDelay = 2 * time.Second
	}

	statuses := cfg.Retry.RetryableStatuses
	if len(statuses) == 0 {
		statuses = []int{
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		}
	}

	methods := cfg.Retry.RetryableHTTPMethods
	if len(methods) == 0 {
		methods = []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodOptions,
			http.MethodPut,
			http.MethodDelete,
		}
	}

	return &Client{
		httpClient:        httpClient,
		maxAttempts:       maxAttempts,
		baseDelay:         baseDelay,
		maxDelay:          maxDelay,
		retryableStatuses: toStatusSet(statuses),
		retryableMethods:  toMethodSet(methods),
	}
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("httpclient: nil request")
	}

	maxAttempts := c.maxAttempts
	if req.Body != nil && req.GetBody == nil {
		// Body를 재생성할 수 없으면 안전하게 단일 시도로 제한한다.
		maxAttempts = 1
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		clonedReq, err := cloneRequest(ctx, req)
		if err != nil {
			return nil, err
		}

		c.setTraceHeaderFromContext(ctx, clonedReq)

		resp, doErr := c.httpClient.Do(clonedReq)
		if !c.shouldRetry(req, resp, doErr, attempt, maxAttempts) {
			return resp, doErr
		}

		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		if err := sleepWithContext(ctx, c.nextDelay(attempt)); err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("httpclient: unexpected retry termination")
}

func (c *Client) setTraceHeaderFromContext(ctx context.Context, req *http.Request) {
	if req.Header == nil {
		req.Header = make(http.Header)
	}

	traceID := kitlog.GetTraceID(ctx)
	if traceID == kitlog.Unknown {
		traceID = kitlog.NewTraceID()
	}
	req.Header.Set(kitlog.TraceHeader, traceID)

	// pSpanID는 현재 spanId와 동일한 값으로 내려준다. 다음 client 입장에서는 parent.
	// 다음 spanID는 새로 생성해서 내려준다.
	req.Header.Set(kitlog.PSpanHeader, kitlog.GetSpanID(ctx))
	req.Header.Set(kitlog.SpanHeader, kitlog.NewSpanID())
}

func (c *Client) shouldRetry(originReq *http.Request, resp *http.Response, err error, attempt, maxAttempts int) bool {
	if attempt >= maxAttempts {
		return false
	}

	if !c.isRetryableMethod(originReq.Method) {
		return false
	}

	if err != nil {
		return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
	}

	if resp == nil {
		return false
	}

	_, ok := c.retryableStatuses[resp.StatusCode]
	return ok
}

func (c *Client) isRetryableMethod(method string) bool {
	_, ok := c.retryableMethods[strings.ToUpper(method)]
	return ok
}

func (c *Client) nextDelay(attempt int) time.Duration {
	delay := c.baseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= c.maxDelay {
			return c.maxDelay
		}
	}
	if delay > c.maxDelay {
		return c.maxDelay
	}
	return delay
}

func cloneRequest(ctx context.Context, req *http.Request) (*http.Request, error) {
	cloned := req.Clone(ctx)

	if req.Body == nil {
		return cloned, nil
	}

	if req.GetBody == nil {
		cloned.Body = req.Body
		return cloned, nil
	}

	body, err := req.GetBody()
	if err != nil {
		return nil, fmt.Errorf("httpclient: clone body: %w", err)
	}

	cloned.Body = body
	return cloned, nil
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func toStatusSet(statuses []int) map[int]struct{} {
	set := make(map[int]struct{}, len(statuses))
	for _, status := range statuses {
		set[status] = struct{}{}
	}
	return set
}

func toMethodSet(methods []string) map[string]struct{} {
	set := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		upper := strings.ToUpper(strings.TrimSpace(method))
		if upper == "" {
			continue
		}
		if !isValidHTTPMethod(upper) {
			continue
		}
		set[upper] = struct{}{}
	}
	return set
}

func isValidHTTPMethod(method string) bool {
	switch method {
	case http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace:
		return true
	default:
		return false
	}
}
