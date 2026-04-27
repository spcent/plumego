# Card 0348: Contract Error Code Constants Consolidation

Priority: P2
State: done
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: contract, router, x/ai, x/pubsub, x/websocket, x/scheduler, x/gateway, x/messaging
Depends On: 0970

## Goal

Add the missing `CodeMethodNotAllowed` constant to `contract/error_codes.go`,
add `CodeBadGateway` for the gateway extension, and replace every hardcoded
error code string literal (`"METHOD_NOT_ALLOWED"`, `"SERVICE_UNAVAILABLE"`,
`"BAD_GATEWAY"`, `"INVALID_REQUEST"`) with the corresponding `contract.Code*`
constant so all callers share one authoritative definition.

## Problem

`contract/error_codes.go` exports canonical error code constants but is missing
entries for two cross-cutting HTTP semantics:

- `"METHOD_NOT_ALLOWED"` — hardcoded as a string literal in 6 separate files:
  `router/dispatch.go:50`, `x/ai/streaming/handler.go:92`,
  `x/pubsub/distributed.go:749`, `x/pubsub/prometheus.go:511`,
  `x/scheduler/admin_http.go:352`, `x/websocket/server.go:183`
- `"BAD_GATEWAY"` — hardcoded in `x/gateway/config.go:195`; no constant exists

Existing constants are also not adopted by callers:

- `CodeUnavailable = "SERVICE_UNAVAILABLE"` — inlined as a raw string in
  `x/gateway/config.go:180`, `x/messaging/api.go:170`,
  `x/scheduler/admin_http.go:36`
- `CodeBadRequest = "BAD_REQUEST"` is present but `"INVALID_REQUEST"` (a
  distinct code used by `x/ai/streaming/handler.go:99` and
  `x/messaging/api.go:23,48`) has no constant yet

This leaves the codebase with a split vocabulary: the constant is defined in one
place but the string is scattered and can drift independently.

## Scope

- Add to `contract/error_codes.go`:
  ```go
  CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
  CodeBadGateway       = "BAD_GATEWAY"
  CodeInvalidRequest   = "INVALID_REQUEST"
  ```
- Replace every hardcoded occurrence of those strings (and existing
  `"SERVICE_UNAVAILABLE"`) with the matching `contract.Code*` constant across
  all callers listed above.
- Keep existing constant names and values unchanged.

## Non-Goals

- Do not rename or renumber existing constants.
- Do not add constants for codes that appear only once inside a narrow
  package-private context (e.g., `"JOIN_DENIED"`, `"TRIGGER_FAILED"`,
  `"SERVICE_UNAVAILABLE"` used inside admin-only scheduler path that is already
  being rewritten in card 0971).
- Do not change error messages, HTTP status codes, or response categories.

## Files

- `contract/error_codes.go`
- `router/dispatch.go`
- `x/ai/streaming/handler.go`
- `x/pubsub/distributed.go`
- `x/pubsub/prometheus.go`
- `x/websocket/server.go`
- `x/scheduler/admin_http.go`
- `x/gateway/config.go`
- `x/messaging/api.go`

## Tests

```bash
rg -n '"METHOD_NOT_ALLOWED"|"BAD_GATEWAY"|"INVALID_REQUEST"|"SERVICE_UNAVAILABLE"' \
  router x/ai x/pubsub x/websocket x/scheduler x/gateway x/messaging -g '*.go' \
  --glob '!*_test.go'
go test -timeout 20s ./contract/... ./router/... ./x/ai/... ./x/pubsub/... ./x/websocket/... ./x/scheduler/... ./x/gateway/... ./x/messaging/...
go vet ./contract/... ./router/... ./x/ai/... ./x/pubsub/... ./x/websocket/... ./x/scheduler/... ./x/gateway/... ./x/messaging/...
```

## Docs Sync

None required — constants are package-internal API, not user-facing docs.

## Done Definition

- `contract/error_codes.go` exports `CodeMethodNotAllowed`, `CodeBadGateway`,
  and `CodeInvalidRequest`.
- The search above returns no matches from the non-test source files.
- All targeted package tests pass.
- `go vet` is clean.

## Outcome

