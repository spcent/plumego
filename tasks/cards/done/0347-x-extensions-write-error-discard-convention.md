# Card 0347: X Extensions WriteError Discard Convention

Priority: P2
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/ops, x/gateway, x/ai, x/tenant, x/devtools, x/frontend, x/resilience, x/fileapi, x/scheduler, x/websocket
Depends On: 0970

## Goal

Apply the canonical `_ = contract.WriteError(...)` discard prefix to every bare
`contract.WriteError` call across all x/* extension packages that were not
covered by card 0970, making the entire codebase consistent with style guide
§17.3.

## Problem

Approximately 40 `contract.WriteError` calls in x/* extension packages omit the
`_ =` discard prefix. Affected files and approximate counts:

- `x/ops/healthhttp/helpers.go` — 2 bare calls
- `x/ops/healthhttp/history.go` — 2 bare calls
- `x/ops/healthhttp/handlers.go` — 1 bare call
- `x/gateway/config.go` — 3 bare calls (lines 178, 185, 193)
- `x/gateway/transform/transform.go` — 1 bare call (line 109)
- `x/ai/sse/sse.go` — 1 bare call (line 256)
- `x/ai/streaming/handler.go` — 7 bare calls
- `x/tenant/session/middleware.go` — 3 bare calls (lines 167, 184, 188)
- `x/devtools/devtools.go` — 1 bare call (line 216)
- `x/frontend/frontend.go` — 2 bare calls (lines 347, 529)
- `x/resilience/circuitbreaker/middleware.go` — 2 bare calls (lines 107, 121)
- `x/fileapi/handler.go` — 1 bare call (line 262)
- `x/scheduler/admin_http.go` — 3 bare calls (lines 36, 94, 168)
- `x/websocket/server.go` — 11 bare calls

Note: bare calls inside `x/ops/ops.go` (7 calls) and `x/messaging/api.go` are
intentionally excluded here because those handlers will be rewritten as part of
the adaptCtx migration (cards 0974–0975), which eliminates the handlers
wholesale rather than patching them in place.

## Scope

- Prefix every bare `contract.WriteError(...)` call in the listed files with
  `_ =`.
- No logic changes, no signature changes, no test content changes.

## Non-Goals

- Do not touch `x/ops/ops.go` or `x/messaging/api.go` — those are covered by
  the adaptCtx migration cards 0974–0975.
- Do not change error builder chains or response content.
- Do not add new tests.

## Files

- `x/ops/healthhttp/helpers.go`
- `x/ops/healthhttp/history.go`
- `x/ops/healthhttp/handlers.go`
- `x/gateway/config.go`
- `x/gateway/transform/transform.go`
- `x/ai/sse/sse.go`
- `x/ai/streaming/handler.go`
- `x/tenant/session/middleware.go`
- `x/devtools/devtools.go`
- `x/frontend/frontend.go`
- `x/resilience/circuitbreaker/middleware.go`
- `x/fileapi/handler.go`
- `x/scheduler/admin_http.go`
- `x/websocket/server.go`

## Tests

```bash
rg -n 'contract\.WriteError' x/ops/healthhttp x/gateway x/ai x/tenant/session x/devtools x/frontend x/resilience x/fileapi x/scheduler x/websocket -g '*.go' | grep -v '_ = contract\.WriteError'
go test -timeout 20s ./x/ops/... ./x/gateway/... ./x/ai/... ./x/tenant/... ./x/devtools/... ./x/frontend/... ./x/resilience/... ./x/fileapi/... ./x/scheduler/... ./x/websocket/...
go vet ./x/ops/... ./x/gateway/... ./x/ai/... ./x/tenant/... ./x/devtools/... ./x/frontend/... ./x/resilience/... ./x/fileapi/... ./x/scheduler/... ./x/websocket/...
```

## Docs Sync

None required — this is a pure style fix.

## Done Definition

- `rg -n 'contract\.WriteError' x/ -g '*.go' | grep -v '_ = '` returns no
  results from the targeted files (x/ops/ops.go and x/messaging/api.go excluded
  pending their own migration).
- All targeted package tests pass.
- `go vet` is clean across all targeted packages.

## Outcome

