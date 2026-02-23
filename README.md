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
	defer kitlog.Sync()

	ctx := kitlog.WithTraceID(context.Background(), "abc-123")
	kitlog.Info(ctx, "user login")
	kitlog.Info(context.Background(), "no trace id in context")
}
```

`context`에 trace id가 없으면 `traceId=unknown`이 자동으로 기록됩니다.
