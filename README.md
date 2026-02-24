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

## gRPC Client + Interceptor (Trace)

```go
package main

import (
	"context"
	"log"

	kitgrpc "github.com/NamhaeSusan/my-go-kit/grpcclient"
	kitlog "github.com/NamhaeSusan/my-go-kit/log"
)

func main() {
	client, err := kitgrpc.NewClient("localhost:50051", kitgrpc.Config{})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := kitlog.WithTraceID(context.Background(), "trace-grpc-1")
	ctx = kitlog.WithSpanID(ctx, "span-grpc-1")

	// client.GetConn().Invoke / generated pb client에 conn 전달 시 interceptor가 metadata를 자동 주입
	_ = client.GetConn()
	_ = ctx
}
```

- client interceptor는 context의 trace/span 값을 gRPC metadata(`x-trace-id`, `x-span-id`, `x-pspan-id`)로 자동 전파합니다.
- server interceptor는 inbound metadata를 context로 복원하며, 누락 값은 trace/span 자동 생성 + pSpan=`unknown`을 사용합니다.
