# 2026-02-23 Gin trace 미들웨어 추가

## 작업 요약
- `middleware` 패키지 신규 추가
- Gin용 traceId 미들웨어 구현 (`GinTraceID`, `GinTraceIDWithConfig`)
- `X-Trace-Id` 헤더가 없으면 UUID 생성 후 request context/response header/Gin context에 주입
- 미들웨어 동작 테스트 4종 추가
- `README.md`, `CLAUDE.md` 문서 업데이트

## 변경된 파일
- `middleware/gin_trace.go` (신규)
- `middleware/gin_trace_test.go` (신규)
- `README.md`
- `CLAUDE.md`
- `go.mod`, `go.sum`

## 검증
- `go test ./...` 통과
