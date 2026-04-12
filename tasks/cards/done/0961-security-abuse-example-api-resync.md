# Card 0961: Security Abuse Example API Resync

Priority: P2
State: done
Primary Module: security

## Goal

Align `security/abuse` package examples with the real stable limiter API so the
package comment teaches the current constructor and return shape instead of a
removed pre-refactor surface.

## Problem

- `security/abuse/limiter.go` package-level example still uses
  `abuse.NewGuard(...)`, which no longer exists.
- The same example also treats `Allow(...)` like a boolean gate, while the real
  API returns a `Decision` value.
- The actual stable constructor path is `NewLimiter(Config)`, and callers are
  expected to inspect `Decision.Allowed` and related fields.

This is a stable learning-surface drift problem: the implementation and tests
already use the canonical API, but the package doc still teaches a stale one.

## Scope

- Rewrite the package-level example in `security/abuse/limiter.go` to use
  `NewLimiter(...)` plus `Decision`.
- Check nearby stable security docs for the same stale `NewGuard` wording and
  fix them if present.
- Keep examples aligned with the current return types and names.

## Non-Goals

- Do not add a `NewGuard(...)` compatibility wrapper.
- Do not change limiter runtime behavior, metrics, or eviction logic in this
  card.
- Do not widen the card into a broader security refactor.

## Files

- `security/abuse/limiter.go`
- `docs/modules/security/README.md`
- any touched stable docs that still mention `NewGuard(...)`

## Tests

- `go test -timeout 20s ./security/abuse`
- `go test -race -timeout 60s ./security/abuse`
- `go vet ./security/abuse`
- `rg -n 'NewGuard\\(|Allow\\(.*\\) \\{' security docs/modules README.md README_CN.md`

## Docs Sync

- No repo-wide doc sync expected beyond touched stable security docs unless the
  stale constructor name appears elsewhere.

## Done Definition

- `security/abuse` package docs no longer reference `NewGuard(...)`.
- Examples use `NewLimiter(...)` and the `Decision` shape correctly.
- Grep for the stale constructor name is empty in touched stable docs.

## Outcome

- Rewrote the `security/abuse` package-level example to use
  `NewLimiter(...)` instead of the removed `NewGuard(...)` constructor.
- Updated the example to consume the returned `Decision` shape via
  `decision.Allowed` instead of treating `Allow(...)` like a boolean.
- Confirmed no nearby stable docs still advertise `NewGuard(...)`.

## Validation Run

```bash
go test -timeout 20s ./security/abuse
go test -race -timeout 60s ./security/abuse
go vet ./security/abuse
rg -n 'NewGuard\(|Allow\(.*\) \{' security docs/modules README.md README_CN.md
```
