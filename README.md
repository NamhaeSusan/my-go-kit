# my-go-kit

`my-go-kit`은 Go 서비스에서 공통으로 쓰는 기반 기능을 모아둔 경량 유틸 패키지입니다.

구성:
- `log`: `zap` 기반 구조화 로깅 + trace/span context 필드
- `middleware`: Gin용 trace/span 전파 미들웨어
- `httpclient`: trace 헤더 전파 + 재시도 HTTP 클라이언트
- `grpcclient`: gRPC 연결 풀 + trace/logging 인터셉터

## Install

```bash
go get github.com/NamhaeSusan/my-go-kit
```

## 1) Logging (`log`)

```go
package main

import (
	"context"

	kitlog "github.com/NamhaeSusan/my-go-kit/log"
)

func main() {
	_ = kitlog.Init("")
	defer kitlog.Close()

	ctx := kitlog.WithTraceID(context.Background(), "trace-1")
	ctx = kitlog.WithSpanID(ctx, "span-1")
	ctx = kitlog.WithPSpanID(ctx, "unknown")

	kitlog.Infof(ctx, "user login: %s", "kim")
}
```

동작:
- context에 값이 없으면 `traceId/spanId/pSpanId`는 `unknown`으로 기록됩니다.
- 로그 필드명: `traceId`, `spanId`, `pSpanId`

## 2) Gin Middleware (`middleware`)

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

동작:
- 요청 헤더 `X-Trace-Id`, `X-Span-Id`, `X-PSpan-Id`를 읽어 request context와 gin context에 주입
- 누락 시:
  - `traceId`: 자동 생성
  - `spanId`: 자동 생성
  - `pSpanId`: `unknown`
- 기본값으로 동일 헤더를 response에도 기록

## 3) HTTP Client (`httpclient`)

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

	ctx := kitlog.WithTraceID(context.Background(), "trace-http-1")
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	_, _ = client.Do(ctx, req)
}
```

동작:
- outbound header 자동 주입:
  - `X-Trace-Id`: context trace 또는 신규 생성
  - `X-PSpan-Id`: 현재 span
  - `X-Span-Id`: 신규 span
- 기본 재시도 메서드: `GET/HEAD/OPTIONS/PUT/DELETE`
- 기본 재시도 상태코드: `429/500/502/503/504`
- 재시도 간격: 지수 백오프 (`BaseDelay` ~ `MaxDelay`)

## 4) gRPC Client (`grpcclient`)

```go
package main

import (
	"log"
	"time"

	kitgrpc "github.com/NamhaeSusan/my-go-kit/grpcclient"
)

func main() {
	client, err := kitgrpc.NewClient("localhost:50051", kitgrpc.Config{
		MaxConnections: 4,
		IdleTimeout:    10 * time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	conn := client.GetConn()
	_ = conn
}
```

클라이언트 기본 인터셉터:
- Trace 전파 인터셉터
- Logging 인터셉터

클라이언트 로깅 필드:
- `elapsed` (ms)
- `method`
- `service`
- `grpc_code`
- `log_type=grpc`

서버 측 인터셉터도 별도 제공:
- `interceptor.UnaryServerTraceInterceptor()`
- `interceptor.StreamServerTraceInterceptor()`
- `interceptor.UnaryServerLoggingInterceptor()`
- `interceptor.StreamServerLoggingInterceptor()`

## 패키지 구조

```text
log/
middleware/
httpclient/
grpcclient/
```
