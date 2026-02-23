# my-go-kit

`zap` 기반의 경량 로깅 키트입니다.

## Quick Start

```go
package main

import (
	"context"

	kitlog "github.com/NamhaeSusan/my-go-kit/log"
)

func main() {
	_ = kitlog.Init("")
	defer kitlog.Close() // stops SIGHUP listener + Sync

	ctx := kitlog.WithTraceID(context.Background(), "abc-123")
	kitlog.Infof(ctx, "user login: %s", "kim")
	kitlog.Infof(context.Background(), "no trace id in context")
}
```

`context`에 trace id가 없으면 `traceId=unknown`이 자동으로 기록됩니다.

## Gin Middleware

```go
package main

import (
	"github.com/gin-gonic/gin"
	kitmw "github.com/NamhaeSusan/my-go-kit/middleware"
)

func main() {
	r := gin.New()
	r.Use(kitmw.GinTraceID())
}
```

`X-Trace-Id`가 없으면 trace id를 자동 생성해 request context와 response header에 주입합니다.

## HTTP Client (Trace + Retry)

```go
package main

import (
	"context"
	"net/http"
	"time"

	kithttp "github.com/NamhaeSusan/my-go-kit/httpclient"
	kitlog "github.com/NamhaeSusan/my-go-kit/log"
)

func main() {
	client := kithttp.New(kithttp.Config{
		Retry: kithttp.RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   100 * time.Millisecond,
			MaxDelay:    1 * time.Second,
		},
	})

	ctx := kitlog.WithTraceID(context.Background(), "trace-abc")
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)
	_, _ = client.Do(ctx, req)
}
```

- request context에 traceId가 있으면 `X-Trace-Id` 헤더로 자동 전파됩니다.
- traceId가 없으면 자동 생성 후 헤더에 주입됩니다.
- 기본 재시도 대상은 `GET/HEAD/OPTIONS/PUT/DELETE` + `429/500/502/503/504` 입니다.
