# Card 0973: DebugErrors Config-Merge and Webhook HTTP Code Cleanup

Priority: P3
State: active
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: middleware/debug, x/webhook
Depends On: —

## Goal

Fix two small but semantically incorrect patterns in adjacent packages:

1. `middleware/debug.DebugErrors` merges the caller-supplied config into a
   hard-coded default via a piecemeal field-by-field guard, which inverts the
   expected override semantics for boolean fields.
2. `x/webhook/in.go` inlines HTTP-layer error code strings (`"missing_secret"`,
   `"invalid_signature"`, `"publish_failed"`) as raw literals instead of
   package-level named constants, creating an undiscoverable second vocabulary
   alongside the already-defined `webhook.ErrorCode` constants.

## Problem

### `middleware/debug/debug_errors.go` lines 95–108

```go
func DebugErrors(config DebugErrorConfig) middleware.Middleware {
    cfg := DefaultDebugErrorConfig()          // defaults: IncludeRequest=true, IncludeQuery=true
    if config.NotFoundHint != "" {
        cfg.NotFoundHint = config.NotFoundHint
    }
    if config.IncludeBody {
        cfg.IncludeBody = true
    }
    if !config.IncludeRequest {
        cfg.IncludeRequest = false
    }
    if !config.IncludeQuery {
        cfg.IncludeQuery = false
    }
```

This merge pattern silently discards `IncludeBody: false` (the caller's
explicit opt-out is ignored because the guard only fires when the value is
`true`) and forces `IncludeRequest` and `IncludeQuery` to default `true` unless
the caller explicitly opts out. A caller passing `DebugErrorConfig{}` (the zero
value) unexpectedly receives `IncludeRequest: true, IncludeQuery: true` even
though they expressed no preference. The fix is `cfg := config`; the default
constructor `DefaultDebugErrorConfig()` is kept as a public convenience for
callers who want the recommended settings.

### `x/webhook/in.go`

The HTTP transport handlers use raw string literals as `contract.NewErrorBuilder`
`.Code(...)` arguments:

- `"missing_secret"` — appears twice (GitHub + Stripe paths)
- `"invalid_signature"` — appears twice
- `"publish_failed"` — appears twice

These should be promoted to unexported package-level constants alongside the
existing `webhook.ErrorCode` constants so a single change point controls all
uses:

```go
const (
    httpCodeMissingSecret   = "missing_secret"
    httpCodeInvalidSignature = "invalid_signature"
    httpCodePublishFailed   = "publish_failed"
)
```

## Scope

- Replace the piecemeal config-merge body in `DebugErrors` with `cfg := config`.
- Add the three `httpCode*` constants to `x/webhook/in.go` (or a new small
  `in_codes.go`) and replace the six inline literals with the constants.
- Adjust or add a focused test that passes `DebugErrorConfig{}` and verifies
  that `IncludeRequest` and `IncludeQuery` are treated as `false` (not
  defaulted to `true`).

## Non-Goals

- Do not change `DefaultDebugErrorConfig` — it remains valid for callers who
  want the recommended development defaults.
- Do not change error messages, HTTP status codes, or categories in either file.
- Do not redesign webhook signature-verification or deduplication logic.

## Files

- `middleware/debug/debug_errors.go`
- `middleware/debug/debug_errors_test.go` (or equivalent test file)
- `x/webhook/in.go`

## Tests

```bash
rg -n '"missing_secret"\|"invalid_signature"\|"publish_failed"' x/webhook/in.go
go test -timeout 20s ./middleware/debug/... ./x/webhook/...
go vet ./middleware/debug/... ./x/webhook/...
```

## Docs Sync

- `docs/modules/middleware-debug/README.md` if it documents the merge behavior
  or shows a zero-value config example.

## Done Definition

- `DebugErrors` uses `cfg := config` instead of the piecemeal field guards.
- A test confirms that `DebugErrorConfig{}` does not auto-enable
  `IncludeRequest` or `IncludeQuery`.
- `x/webhook/in.go` uses named constants for its three HTTP-layer error codes.
- All tests pass; `go vet` is clean.

## Outcome

