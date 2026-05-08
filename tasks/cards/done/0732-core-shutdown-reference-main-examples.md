# 0732 - core Shutdown Reference Main Examples

State: done
Priority: P2
Primary Module: core

## Goal

Finish shutdown example cleanup for reference `main.go` files that still defer
`Shutdown` without handling or explicitly discarding the returned error.

## Scope

- Update remaining reference `main.go` direct shutdown defers.
- Keep example control flow and startup behavior unchanged.
- Use explicit `_ =` discards where there is no useful error recovery path.

## Non-goals

- Do not change route wiring or extension behavior.
- Do not add dependencies or logging surfaces.
- Do not touch already-updated reference app packages.

## Files

- `reference/with-ai/main.go`
- `reference/with-rest/main.go`
- `reference/with-tenant/main.go`

## Tests

- `go test -timeout 20s ./reference/...`
- `go run ./internal/checks/reference-layout`

## Docs Sync

Not required.

## Done Definition

- No reference `defer app.Shutdown(...)` call silently ignores the returned
  error.
- Reference tests and layout checks pass.

## Outcome

- Updated `with-ai`, `with-rest`, and `with-tenant` reference main examples to
  log deferred shutdown errors.
- Verified with `go test -timeout 20s ./reference/...`,
  `go run ./internal/checks/reference-layout`, and an `rg` scan for remaining
  direct deferred shutdown calls.
