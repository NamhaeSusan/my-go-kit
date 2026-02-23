# 2026-02-23 Code Review & Test Fix

## 작업 요약
전체 코드 리뷰 수행 및 실패하는 테스트 3건 수정 + 리뷰에서 발견된 추가 버그 수정.

## 수정된 버그

### 1. httpclient/client.go — ctx 파라미터 미사용 버그 (3곳)
- `Do()` 메서드가 `ctx` 파라미터 대신 `req.Context()`를 사용하는 곳이 3곳
  - `setTraceHeaderFromContext(req.Context(), ...)` → `ctx` 사용
  - `sleepWithContext(req.Context(), ...)` → `ctx` 사용 (retry 사이 취소 동작 불가)
  - `cloneRequest(req)` → `cloneRequest(ctx, req)` (HTTP 요청 취소 동작 불가)

### 2. httpclient/client.go — trace 자동 생성 누락
- context에 traceId/spanId가 없을 때 헤더를 아예 설정하지 않음
- 미들웨어와 동일하게 자동 생성 로직 추가

### 3. httpclient/client.go — cloneRequest 중복 헤더 클론
- `req.Clone()`이 이미 헤더를 deep copy하는데, `req.Header.Clone()`으로 중복 복사
- 불필요한 할당 제거

### 4. httpclient/client_test.go — span 전파 테스트 기대값 불일치
- 설계 의도: 현재 spanId → pSpanId, 새 spanId 생성하여 downstream 전달
- 테스트가 원본 spanId를 그대로 기대하고 있었음

### 5. middleware/gin_trace_test.go — spanId 길이 검증 오류
- isHex32(32자)로 spanId(16자) 검증 → isHex16 헬퍼 추가

### 6. log/logger.go — MaxSize 주석 오류
- `MaxSize: 1024, // 1MB` → lumberjack MaxSize는 MB 단위이므로 1GB가 맞음

## 코드 리뷰 추가 발견사항 (미수정, 향후 고려)
- `log/logger.go`: os.MkdirAll에 os.ModePerm(0777) 사용 → 0o750 권장
- `log/logger.go`: SIGHUP 시그널 핸들러 goroutine 누수 (stop 불가)
- `log/logger.go`: sync.Once로 Init 실패 시 재초기화 불가
- `httpclient/client.go`: http.DefaultClient 직접 참조 (공유 상태 변이 위험)
- context key 직접 참조 대신 Get 함수 사용 권장 (캡슐화)
- trace 헤더 상수 중복 정의 (middleware, httpclient)
- CLAUDE.md에 GenerateFunc 커스터마이징 언급하나 실제 구현 없음

## 변경 파일
- `httpclient/client.go` — ctx 사용, 자동 생성, 중복 클론 제거
- `httpclient/client_test.go` — 테스트 기대값 수정
- `middleware/gin_trace_test.go` — isHex16 헬퍼 추가
- `log/logger.go` — MaxSize 주석 수정

## 검증
- `make test` — 전체 통과
- `make lint` — 이슈 0건
