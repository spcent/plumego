# Card 0970: Stable Roots WriteError Discard Convention

Priority: P1
State: active
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: middleware, router
Depends On: —

## Goal

Apply the canonical `_ = contract.WriteError(...)` discard prefix to every bare
`contract.WriteError` call in the stable root packages so the transport layer
obeys the style rule that `WriteError`'s error return is always explicitly
discarded (style guide §17.3).

## Problem

Eight `contract.WriteError` calls in stable-root packages omit the `_ =` discard
prefix, violating the canonical pattern and producing spurious staticcheck/errcheck
warnings:

- `middleware/auth/contract.go` lines 129, 143, 151 — three bare calls in
  `WriteAuthError` and `ApplyAuthMiddleware`
- `middleware/error_registry.go` line 27 — bare call inside
  `WriteTransportError`
- `middleware/timeout/timeout.go` lines 96, 182 — two bare calls in the
  timeout middleware response writers
- `middleware/debug/debug_errors.go` line 129 — bare call in `DebugErrors`
  intercept path
- `router/dispatch.go` line 48 — bare call in the method-not-allowed dispatch
  branch

Because these are stable roots (imported by downstream extensions and user apps),
this pattern propagates as a teaching example into every package that reads the
source.

## Scope

- Prefix every bare `contract.WriteError(...)` in the files listed above with
  `_ =`.
- No logic changes, no signature changes, no test renames.

## Non-Goals

- Do not fix x/* extension packages (see card 0971).
- Do not add new tests beyond what already exercise these paths.
- Do not change error builder chains or error content.

## Files

- `middleware/auth/contract.go`
- `middleware/error_registry.go`
- `middleware/timeout/timeout.go`
- `middleware/debug/debug_errors.go`
- `router/dispatch.go`

## Tests

```bash
rg -n 'contract\.WriteError' middleware router -g '*.go' | grep -v '_ = contract\.WriteError'
go test -timeout 20s ./middleware/... ./router/...
go vet ./middleware/... ./router/...
```

## Docs Sync

None required — this is a pure style fix.

## Done Definition

- `rg -n 'contract\.WriteError' middleware router -g '*.go' | grep -v '_ = '`
  returns no results from the targeted files.
- All middleware and router tests pass.
- `go vet ./middleware/... ./router/...` is clean.

## Outcome

