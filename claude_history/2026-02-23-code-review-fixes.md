# Code Review Fixes

## Summary

코드 리뷰에서 발견된 HIGH/MEDIUM 이슈 7건을 수정했다.

## Changes

- **Unknown export** — `log/context.go`의 `"unknown"` 하드코딩을 `Unknown` 상수로 export
- **Header constants centralized** — `TraceHeader`, `SpanHeader`, `PSpanHeader` 상수를 `log/context.go`에 정의하고 `middleware`, `httpclient`에서 참조
- **os.MkdirAll 0o750** — `log/logger.go`의 로그 디렉토리 생성 권한을 `0o750`으로 강화
- **Close() added** — `log/logger.go`에 `Close()` 함수 추가 (SIGHUP 리스너 정리 + Sync)
- **http.DefaultClient removed** — `httpclient/client.go`에서 `http.DefaultClient` 사용 제거
- **Public API usage in httpclient** — `httpclient`가 `log` 패키지의 public API(`Unknown`, header 상수)를 사용하도록 변경
- **GenerateFunc removed** — `middleware/gin_trace.go`의 `GinTraceIDWithConfig`에서 `GenerateFunc` 제거

## Changed Files

- `log/context.go`
- `log/context_test.go`
- `log/logger.go`
- `log/logger_test.go`
- `httpclient/client.go`
- `httpclient/client_test.go`
- `middleware/gin_trace.go`
- `middleware/gin_trace_test.go`
- `CLAUDE.md`
- `README.md`

## Verification

- `make test` — PASS
- `make lint` — PASS
