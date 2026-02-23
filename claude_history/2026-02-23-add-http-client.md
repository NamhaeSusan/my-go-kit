# 2026-02-23 HTTP 클라이언트 추가

## 작업 요약
- `httpclient` 패키지 신규 추가
- trace 호환: request context의 traceId를 `X-Trace-Id` 헤더로 자동 주입
- context에 traceId가 없을 경우 UUID 기반 traceId 자동 생성
- retry config 추가: 최대 시도/백오프/재시도 상태코드/재시도 메서드 설정
- 요청 body 재생성 불가(`GetBody == nil`) 시 안전하게 단일 시도로 제한
- README.md, CLAUDE.md 문서 업데이트

## 변경된 파일
- `httpclient/client.go` (신규)
- `httpclient/client_test.go` (신규)
- `README.md`
- `CLAUDE.md`
- `claude_history/2026-02-23-add-http-client.md` (신규)

## 검증
- `go test ./...` 통과
- `make lint` 통과
