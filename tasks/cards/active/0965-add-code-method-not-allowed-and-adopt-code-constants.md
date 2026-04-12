# Card 0965: Add CodeMethodNotAllowed and Adopt contract.Code* Constants Throughout

Milestone:
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P2
State: active
Primary Module: contract
Owned Files: contract/error_codes.go, router/dispatch.go, x/scheduler/admin_http.go, x/pubsub/distributed.go, x/websocket/server.go, x/ai/streaming/handler.go, x/gateway/config.go, x/messaging/api.go
Depends On:

## Goal

Add a missing `CodeMethodNotAllowed` constant to `contract/error_codes.go` and replace
all inline error-code string literals in production code with their `contract.Code*`
equivalents. This eliminates silent typo risk, makes code-search reliable (`rg CodeMethodNotAllowed`
finds all method-not-allowed sites), and aligns with the pattern established by card 0938
(which did this for test files).

## Problem

`contract/error_codes.go` defines uppercase constants for every canonical error code
(`CodeInternalError`, `CodeUnavailable`, `CodeTimeout`, …) **except** `METHOD_NOT_ALLOWED`.
Yet this code appears as a hardcoded string literal in four extension packages and in the
router itself:

| File | Inline literal | Should use |
|---|---|---|
| `router/dispatch.go:50` | `"METHOD_NOT_ALLOWED"` | `contract.CodeMethodNotAllowed` (new) |
| `x/scheduler/admin_http.go:359` | `"METHOD_NOT_ALLOWED"` | `contract.CodeMethodNotAllowed` (new) |
| `x/pubsub/distributed.go` | `"METHOD_NOT_ALLOWED"` | `contract.CodeMethodNotAllowed` (new) |
| `x/websocket/server.go` | `"METHOD_NOT_ALLOWED"` | `contract.CodeMethodNotAllowed` (new) |
| `x/ai/streaming/handler.go` | `"METHOD_NOT_ALLOWED"` | `contract.CodeMethodNotAllowed` (new) |

Additionally, three files inline `"SERVICE_UNAVAILABLE"` instead of using the already-defined
`contract.CodeUnavailable`:

| File | Inline literal | Should use |
|---|---|---|
| `x/scheduler/admin_http.go:37` | `"SERVICE_UNAVAILABLE"` | `contract.CodeUnavailable` |
| `x/gateway/config.go` | `"SERVICE_UNAVAILABLE"` | `contract.CodeUnavailable` |
| `x/messaging/api.go` | `"SERVICE_UNAVAILABLE"` | `contract.CodeUnavailable` |

Card 0938 cleaned up test files; this card cleans up production files.

Note: `x/scheduler/admin_http.go:152` uses `"TRIGGER_FAILED"` which is a scheduler-specific
code with no canonical equivalent. It should remain as a package-local constant
(`triggerFailedCode = "TRIGGER_FAILED"`) defined at the top of admin_http.go instead of
an inline string literal.

## Scope

1. Add `CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"` to `contract/error_codes.go` (after
   `CodeUnavailable` in the System section).
2. Replace the five `"METHOD_NOT_ALLOWED"` inline strings with `contract.CodeMethodNotAllowed`.
3. Replace the three `"SERVICE_UNAVAILABLE"` inline strings with `contract.CodeUnavailable`.
4. In `x/scheduler/admin_http.go`, extract `"TRIGGER_FAILED"` to a package-level constant
   `const triggerFailedCode = "TRIGGER_FAILED"` and use it in `handleJob`.

## Non-Goals

- Do not change HTTP status codes or error categories.
- Do not rename existing `contract.Code*` constants.
- Do not add constants for one-off scheduler or domain-specific codes beyond the ones
  listed (they can stay as package-local constants).
- Do not touch `x/rest/resource.go`'s `"BATCH_TOO_LARGE"` / `"PROCESSING_ERROR"` — those
  are internal batch-result codes handled separately.

## Files

- `contract/error_codes.go` (add constant)
- `router/dispatch.go`
- `x/scheduler/admin_http.go`
- `x/pubsub/distributed.go`
- `x/websocket/server.go`
- `x/ai/streaming/handler.go`
- `x/gateway/config.go`
- `x/messaging/api.go`

## Tests

```bash
go test -timeout 20s ./contract ./router/... ./x/scheduler/... ./x/pubsub/... ./x/websocket/... ./x/gateway/... ./x/messaging/...
go vet ./contract ./router/... ./x/scheduler/... ./x/pubsub/... ./x/websocket/... ./x/gateway/... ./x/messaging/...
```

Check that zero inline `"METHOD_NOT_ALLOWED"` literals remain in non-test production code:

```bash
rg '"METHOD_NOT_ALLOWED"' --include='*.go' -l | grep -v '_test.go'
```

Check that zero inline `"SERVICE_UNAVAILABLE"` remain except inside `contract/error_codes.go`:

```bash
rg '"SERVICE_UNAVAILABLE"' --include='*.go' -l | grep -v '_test.go' | grep -v 'error_codes.go'
```

## Docs Sync

None.

## Done Definition

- `contract.CodeMethodNotAllowed` is exported from `contract/error_codes.go`.
- Zero bare `"METHOD_NOT_ALLOWED"` string literals in non-test Go files (except the
  constant value itself in `error_codes.go`).
- Zero bare `"SERVICE_UNAVAILABLE"` string literals outside `error_codes.go` in
  non-test Go files.
- `x/scheduler/admin_http.go` uses `triggerFailedCode` package constant for `TRIGGER_FAILED`.
- All targeted tests and vet pass.

## Outcome

