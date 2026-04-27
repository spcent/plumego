# Card 0398: Deduplicate Internal Response Recording Helpers

Priority: P2
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: internal
Owned Files:
- internal/httputil/http_response.go
- internal/httputil/http_response_test.go
- internal/nethttp/response_recorder.go
- middleware/internal/transport/http.go
- middleware/internal/transport/transport_test.go
Depends On:

## Goal

Response recording and minimal response hardening are implemented in multiple
places:

- `internal/httputil/http_response.go` owns `EnsureNoSniff` and `SafeWrite`.
- `internal/nethttp/response_recorder.go` has a response recorder that imports
  `internal/httputil`.
- `middleware/internal/transport/http.go` has a second response recorder plus
  another copy of `EnsureNoSniff` and `SafeWrite`.

The implementations are close but not identical: middleware's recorder tracks
bytes and supports `Hijack`/`Flush`, while `internal/nethttp` has a smaller
copy.  This makes future security/header changes easy to apply in one helper
and miss in the other.

## Scope

- Move the canonical response recorder and no-sniff/safe-write behavior into a
  single internal helper location that both current call sites can use.
- Preserve the public surface of `internal/nethttp.NewResponseRecorder` for
  current callers such as `x/gateway/cache`.
- Preserve middleware behavior, including byte accounting, `Hijack`, and
  `Flush` forwarding.
- Keep the helper internal to the main module; do not expose it as a stable root
  package.

## Non-goals

- Do not change `contract.WriteResponse` or `contract.WriteError`.
- Do not change response caching semantics in `x/gateway/cache`.
- Do not broaden middleware internal packages into public API.
- Do not add dependencies.

## Files

- `internal/httputil/http_response.go`
- `internal/httputil/http_response_test.go`
- `internal/nethttp/response_recorder.go`
- `middleware/internal/transport/http.go`
- `middleware/internal/transport/transport_test.go`

## Tests

```bash
go test -race -timeout 60s ./internal/httputil/... ./internal/nethttp/... ./middleware/...
go test -timeout 20s ./internal/httputil/... ./internal/nethttp/... ./middleware/...
go vet ./internal/httputil/... ./internal/nethttp/... ./middleware/...
```

## Docs Sync

No public docs sync required unless the implementation changes middleware or
gateway observable behavior.

## Done Definition

- There is one canonical implementation for no-sniff safe writes and response
  recorder state transitions.
- `internal/nethttp.NewResponseRecorder` and middleware recorder callers still
  compile without behavior regressions.
- Tests cover status defaulting, header forwarding, bytes written, body capture,
  and repeated `WriteHeader` handling.
- The listed validation commands pass.

## Outcome

- Moved the canonical `ResponseRecorder`, header copy, `EnsureNoSniff`, and `SafeWrite` behavior into `internal/httputil`.
- Kept `internal/nethttp.NewResponseRecorder` as a compatibility constructor backed by the shared implementation.
- Kept middleware transport helpers as thin wrappers/aliases backed by the shared implementation, while preserving bytes-written accounting, body capture, `Hijack`, and `Flush`.
- Updated middleware buffered response flushing to use the shared header-copy helper.
- Added direct `internal/httputil` tests for status defaulting, header forwarding, bytes written, body capture, and repeated `WriteHeader`.
- Validation passed:
  - `go test -race -timeout 60s ./internal/httputil/... ./internal/nethttp/... ./middleware/...`
  - `go test -timeout 20s ./internal/httputil/... ./internal/nethttp/... ./middleware/...`
  - `go vet ./internal/httputil/... ./internal/nethttp/... ./middleware/...`
