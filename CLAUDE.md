# wasmflux-go

Go + WebAssembly 고성능 프레임워크. 60Hz 이상 실시간 데이터 스트림 처리용.

## 구조

- `app.go`, `module.go`, `app_options.go`, `registry.go` — App 라이프사이클, 모듈 DI
- `bridge/` — Go↔JS 인터옵, Encode/Decode (리플렉션), 타입 안전 인자 검증, 콜백 풀링, Promise
- `event/` — 토픽 기반 이벤트 버스 (`On`, `Once`, `Topics`)
- `flux/` — Generic Flux 상태 관리, delta 구독, 미들웨어
- `log/` — 구조화 로거 (Text/JSON 포맷, caller, rate limit)
- `pool/` — Generic 오브젝트 풀 (sync.Pool)
- `ring/` — 링 버퍼 (고빈도 스트리밍 데이터)
- `batch/` — 배치 프로세서
- `tick/` — RAF 루프, interval/timeout 스케줄러
- `errors/` — WASM 컨텍스트 에러 처리, 스택 트레이스 포함 패닉 복구
- `util/` — Debounce, Throttle, Retry, RateLimiter, Goroutine Group
- `internal/jsutil/` — DOM 조작, Fetch/HTTP, LocalStorage/SessionStorage, TypedArray
- `internal/debug/` — 디버그 전용 도구 (build tag: debug)
- `example/` — Go 모듈 + Vite React 예제 (counter, signal, compute)
- `cmd/devserver/` — 개발 HTTP 서버

## 빌드

```bash
make setup         # wasm_exec.js 복사
make build         # WASM 빌드
make build-debug   # debug 태그 포함 빌드
make serve         # 개발 서버 실행
make test          # 테스트 (race detector)
make bench         # 벤치마크
make coverage      # 커버리지 리포트
make check         # fmt + vet + lint + test
make watch         # 파일 변경 감지 빌드
make size          # WASM 바이너리 크기 확인
make setup-example # example용 wasm_exec.js + glue.js 복사
make build-example # example WASM 빌드
make example       # WASM 빌드 + Vite 서버 실행
```

## 모듈 DI (의존성 주입)

```go
// Init에서 서비스 등록
ctx.Provide("counter.store", m.store)

// Start에서 서비스 주입 (모든 Init 완료 후)
store, ok := wasmflux.InjectAs[*flux.Store[State]](m.ctx, "counter.store")
```

- `Provide()` → Init 시점에 호출
- `Inject()` / `InjectAs[T]()` → Start 시점에 호출
- DI는 Init/Start 시 한 번만 호출되므로 60Hz 핫패스 성능에 영향 없음

## 벤치마크 (Apple M4 Pro)

### 핵심 패키지

| 패키지 | 작업 | 속도 | Alloc |
|--------|------|------|-------|
| ring/Buffer | Write | 2.8 ns/op | 0 |
| ring/Buffer | WriteBatch (60개) | 6.7 ns/op | 0 |
| event/Bus | Emit | 7.3 ns/op | 0 |
| flux/Store | Dispatch | 7.4 ns/op | 0 |
| batch/Processor | Push | 3.1 ns/op | 0 |
| pool/ByteBuffer | Get/Put | 7.2 ns/op | 0 |
| log/Logger | Info (2 fields) | 84 ns/op | 2 allocs |
| log/Logger | Debug (skipped) | 0.29 ns/op | 0 |

### DI 오버헤드

| 작업 | 속도 | 비교 |
|------|------|------|
| 직접 필드 접근 | 0.23 ns/op | baseline |
| Registry.Inject | 6.3 ns/op | ~27x (Init/Start에서만 호출) |
| Inject + type assert | 6.2 ns/op | Inject과 동일 |
| Inject (100개 서비스) | 7.2 ns/op | 서비스 수에 무관 |

## 컨벤션

- WASM 전용 코드: `//go:build js && wasm` 빌드 태그
- 순수 Go 패키지 (event, flux, log, pool, ring, batch, util, registry): 빌드 태그 없음 → native 테스트 가능
- Functional options 패턴 (`WithXxx`)
- Generic 타입 (`Pool[T]`, `Store[S]`, `Buffer[T]`)
- 성능 핫패스에서 allocation 최소화
- `internal/` 패키지는 외부 노출 금지
- 디버그 기능: `//go:build debug` 태그로 릴리스에서 제거
- Godoc: exported 타입/함수 주석은 이름으로 시작 (`// Buffer is a generic ring buffer.`)
- Logger: `log.String()`, `log.Int()`, `log.Float()` 필드 타입 사용
- Bridge 인자: `bridge.ArgString()`, `bridge.ArgInt()` 등 타입 안전 접근
- Bridge 변환: `bridge.Encode(struct)` / `bridge.Decode(jsValue, &struct)` — JSON 경유 없음
- 모듈 DI: Init에서 `Provide`, Start에서 `Inject` (등록 순서 = DI 순서)
