# 2026-02-23 로그 킷 기능 확장

## 작업 요약
- spanId/pSpanId context 전파 추가
- lumberjack 파일 로그 로테이션 추가 (SIGHUP 시그널 지원)
- 로그 함수 API 변경: Debug/Info/Warn/Error → Debugf/Infof/Warnf/Errorf (fmt.Sprintf 포맷팅)
- zap.ReplaceGlobals 기반 글로벌 로거로 전환
- CLAUDE.md, Makefile 추가
- golangci-lint 이슈 0개 확인

## 변경된 파일
- `CLAUDE.md` (신규)
- `Makefile` (신규)
- `README.md`
- `go.mod`, `go.sum`
- `log/context.go`, `log/context_test.go`
- `log/logger.go`, `log/logger_test.go`

## 검증
- `make lint` 통과 (0 issues)
