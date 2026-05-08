# 0746 - core Server Override Regression

State: done
Priority: P1
Primary Module: core

## Goal

Add regression coverage for `Server()` returning a caller-mutable
`*http.Server` while preserving core's default installed hooks.

## Scope

- Prove prepared servers install the core handler and connection tracker hooks.
- Prove caller replacement of `Handler` and `ConnState` is caller-owned and can
  bypass core behavior, matching documented `net/http` compatibility.
- Keep the behavior documented as a compatibility edge instead of adding guard
  wrappers.

## Non-goals

- Do not wrap or proxy `*http.Server`.
- Do not make `Server()` return an immutable copy.
- Do not change TLS or HTTP/2 runtime behavior.

## Files

- `core/lifecycle_test.go`

## Tests

- `go test -timeout 20s ./core/...`
- `go vet ./core/...`
- `go run ./internal/checks/dependency-rules`

## Docs Sync

Not required; the ownership warning is already documented.

## Done Definition

- Regression tests cover default installed server hooks and caller-owned
  override behavior.
- Core tests, vet, and dependency boundary checks pass.

## Outcome

- Added regression coverage for the default prepared server handler, connection
  tracker hook, and HTTP/2 disable override.
- Added explicit caller-owned override assertions for `Handler` and `ConnState`.
- Verified with `go test -timeout 20s ./core/...`, `go vet ./core/...`, and
  `go run ./internal/checks/dependency-rules`.
