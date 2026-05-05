# 0743 - core TLS Test Readiness

State: active
Priority: P2
Primary Module: core

## Goal

Make core TLS lifecycle test startup synchronization match the HTTP lifecycle
readiness probe pattern.

## Scope

- Reuse or extend the readiness helper for HTTPS clients.
- Replace direct post-goroutine TLS request with bounded readiness polling.
- Keep runtime behavior unchanged.

## Non-goals

- Do not change TLS preparation behavior.
- Do not alter generated test certificate content.
- Do not add external dependencies.

## Files

- `core/lifecycle_test.go`

## Tests

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`

## Docs Sync

Not required.

## Done Definition

- TLS lifecycle test no longer assumes the server is immediately accepting
  connections after goroutine start.
- Normal and race core tests pass.

