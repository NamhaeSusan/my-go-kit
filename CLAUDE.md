# my-go-kit

## 컨셉

Go 프로젝트에서 반복적으로 사용하는 유틸리티를 모아두는 개인 Go 킷. 현재는 `zap` 기반의 경량 로깅 모듈을 제공한다.

---

## 아키텍처

```
my-go-kit/
├── Makefile                # 빌드/테스트/린트 타겟 (make test, make lint 등)
├── go.mod                  # 모듈 정의 (github.com/NamhaeSusan/my-go-kit)
├── go.sum                  # 의존성 체크섬
├── log/                    # 로깅 패키지
│   ├── context.go          # context 기반 traceId/spanId/pSpanId 전파
│   ├── context_test.go     # context 유닛 테스트
│   ├── logger.go           # zap 로거 초기화 및 레벨별 로그 함수
│   └── logger_test.go      # 로거 유닛 테스트
├── claude_history/         # 작업 기록
└── README.md               # 프로젝트 소개 및 Quick Start
```

---

### 작업 기록 (CRITICAL)

모든 작업이 끝나면 `claude_history/` 폴더에 기록을 남긴다.

- **파일명**: `yyyy-mm-dd-{{work}}.md` (예: `2026-02-23-add-error-kit.md`)
- **내용**: 작업 요약, 변경된 파일, 검증 결과를 간단하게 기록
- 같은 날 여러 작업 시 work 부분으로 구분

---

### 기능 구현 후 필수 체크리스트 (CRITICAL — 절대 빠뜨리지 말 것)

기능을 추가하거나 변경한 뒤에는 **반드시** 아래 문서들을 모두 업데이트해야 한다.
코드 변경만 하고 문서를 빠뜨리면 안 된다. 기능 구현이 끝났다고 판단하기 전에 이 체크리스트를 확인할 것.
CLAUDE.md와 README.md 는 필수로 업데이트 하도록 한다.

| 변경 유형 | 업데이트할 문서 |
|-----------|----------------|
| 새 모듈/파일 추가 | `CLAUDE.md` 아키텍처 트리, `README.md` |
| 새 의존성 추가 | `CLAUDE.md` 기술 스택, `README.md` |
| 새 Feature 추가 | `CLAUDE.md` 기능 상세, `README.md` Features |
| 새 로그 레벨/함수 추가 | `CLAUDE.md` 기능 상세, `README.md` Quick Start |

---

### 작업 방식 (CRITICAL)

**간단한 작업**(단일 파일 수정, 오타 수정, 한 줄짜리 버그 픽스)은 직접 처리해도 된다.

**그 외 모든 작업**은 반드시 **TeamCreate로 에이전트 팀을 구성**해서 병렬로 진행한다:
- 기능 구현 → 구현 에이전트 (`general-purpose`)
- 테스트/검증 → 검증 에이전트 (`tdd-guide`)
- 문서 업데이트 → 문서 에이전트 (`doc-updater`)
- 코드 리뷰 → `code-reviewer` 에이전트

예시 팀 구성:
```
Phase 1 (병렬): 구현 에이전트들 → 코드 구현
Phase 2 (병렬): 검증 에이전트 + code-reviewer → 검증
Phase 3: 문서 에이전트 → 문서 반영
```

---

### 핵심 설계 원칙

1. **Context 기반 추적** — 모든 로그 함수가 `context.Context`를 첫 인자로 받아 traceId/spanId/pSpanId를 자동 전파
2. **글로벌 싱글톤** — `zap.ReplaceGlobals`로 전역 로거를 한 번 초기화, 어디서든 import만으로 사용
3. **제로 설정 기본값** — `Init("")`만으로 콘솔 JSON 출력 시작, 파일 경로 지정 시 파일 출력 자동 추가
4. **안전한 폴백** — context에 traceId가 없으면 `"unknown"` 자동 기록, nil context도 패닉 없이 처리

---

## 기술 스택

| 영역 | 라이브러리 | 용도 |
|------|-----------|------|
| 로깅 | `go.uber.org/zap` | 구조화 로거 (JSON 출력) |
| 로그 로테이션 | `gopkg.in/natefinch/lumberjack.v2` | 파일 로그 자동 로테이션 (크기/기간 기반) |
| 언어 | Go 1.26 | 런타임 |

---

## 핵심 기능 상세

### 로거 초기화 (`log.Init`)
- 콘솔(stdout) JSON 출력 기본 제공
- 파일 경로 지정 시 lumberjack 기반 파일 로테이션 자동 추가 (1GB, 7일, gzip 압축)
- SIGHUP 시그널 수신 시 로그 파일 수동 로테이션
- `sync.Once`로 중복 초기화 방지

### 레벨별 로그 함수
- `Debug`, `Info`, `Warn`, `Error` — context에서 추적 필드를 자동 추출하여 로그에 포함
- `fmt.Sprintf` 기반 메시지 포맷팅 지원 (인자 없으면 포맷팅 생략하는 fast path)

### Context 전파 (`log/context.go`)
- `WithTraceID` / `GetTraceID` — traceId context 주입/추출
- `WithSpanID` / `GetSpanID` — spanId context 주입/추출
- `WithPSpanID` / `GetPSpanID` — pSpanId (부모 span) context 주입/추출
- 미설정 또는 빈 문자열 시 `"unknown"` 폴백

---

## 로드맵

### Phase 1 — 로깅 킷 안정화
- [x] zap 기반 로거 초기화 (콘솔 + 파일)
- [x] context 기반 traceId/spanId/pSpanId 전파
- [x] lumberjack 로그 로테이션
- [ ] Makefile 추가 (build, test, lint 타겟)
- [ ] golangci-lint 설정 추가

### Phase 2 — 유틸리티 확장
- [ ] 에러 핸들링 유틸리티 (errors 패키지)
- [ ] HTTP 미들웨어 (Gin용 traceId 자동 주입)
- [ ] gRPC 인터셉터 (traceId 전파)

### Phase 3 — 운영 도구
- [ ] 헬스체크 유틸리티
- [ ] 메트릭 수집 헬퍼
