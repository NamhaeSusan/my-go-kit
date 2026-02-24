# 2026-02-24 gRPC client/interceptor 추가

## 작업 요약
- `grpcclient` 패키지 신규 추가
- gRPC client용 unary/stream trace interceptor 구현
- gRPC server용 unary/stream trace interceptor 구현
- `NewConn` Dial 래퍼로 기본 trace interceptor 체인 구성 지원
- README/CLAUDE 문서에 gRPC 기능 반영

## 변경 파일
- `grpcclient/client.go` (신규)
- `grpcclient/server_interceptor.go` (신규)
- `grpcclient/client_test.go` (신규)
- `grpcclient/server_interceptor_test.go` (신규)
- `go.mod`
- `go.sum`
- `README.md`
- `CLAUDE.md`

## 검증 결과
- `go test ./...` 통과
