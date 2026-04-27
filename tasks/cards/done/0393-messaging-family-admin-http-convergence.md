# Card 0393: Converge Messaging-Family Admin HTTP Surfaces

Priority: P1
State: done
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Primary Module: x/messaging
Owned Files:
- x/messaging/api.go
- x/pubsub/distributed.go
- x/scheduler/admin_http.go
- x/webhook/out.go
- docs/modules/x-messaging/README.md
Depends On: 2101

## Goal

The messaging family still has several app-facing or admin HTTP surfaces that
do not share one canonical response and error path:

- `x/pubsub/distributed.go` uses `contract.WriteJSON` for cluster health,
  heartbeat, publish, and sync responses.
- `x/scheduler/admin_http.go` uses `contract.WriteJSON` for health, stats,
  jobs, DLQ, and action responses, and has a literal `"TRIGGER_FAILED"` code.
- `x/webhook/out.go` consistently uses `WriteResponse` for success, but many
  error messages pass `err.Error()` directly to clients.
- `x/messaging/api.go` has multiple `err.Error()` client messages and mixes
  low-level service errors with transport responses.

This creates repeated one-off HTTP behavior across subordinate packages that
are meant to be governed through the `x/messaging` family.

## Scope

- Introduce package-local error mapping helpers where needed so service errors
  become stable transport codes/messages.
- Replace direct `contract.WriteJSON` success responses in production HTTP
  handlers with `contract.WriteResponse`, unless a bodyless status response is
  intentionally documented.
- Replace literal error codes such as `"TRIGGER_FAILED"` with existing
  `contract` codes or package-local constants.
- Preserve fail-closed auth behavior for cluster and webhook trigger endpoints.
- Add tests for canonical envelopes and sanitized error messages across the
  changed handlers.

## Non-goals

- Do not merge `x/pubsub`, `x/scheduler`, or `x/webhook` into `x/messaging`.
- Do not change broker, scheduler, or delivery semantics.
- Do not alter webhook signature verification or trigger-token comparison.
- Do not add external dependencies.

## Files

- `x/messaging/api.go`
- `x/pubsub/distributed.go`
- `x/scheduler/admin_http.go`
- `x/webhook/out.go`
- `docs/modules/x-messaging/README.md`

## Tests

```bash
go test -race -timeout 60s ./x/messaging/... ./x/pubsub/... ./x/scheduler/... ./x/webhook/...
go test -timeout 20s ./x/messaging/... ./x/pubsub/... ./x/scheduler/... ./x/webhook/...
go vet ./x/messaging/... ./x/pubsub/... ./x/scheduler/... ./x/webhook/...
```

## Docs Sync

Update `docs/modules/x-messaging/README.md` only if the public admin response
shape changes.  The doc should say that app-facing JSON handlers use
`contract.WriteResponse` and errors use stable transport codes.

## Done Definition

- Production HTTP handlers in the listed files no longer use
  `contract.WriteJSON` for JSON success bodies.
- No changed handler returns raw internal `err.Error()` text for server-side
  failures.
- Tests cover at least one success envelope and one sanitized error per changed
  subordinate surface.
- The listed validation commands pass.

## Outcome

- Replaced production `contract.WriteJSON` success bodies in the changed messaging-family admin HTTP handlers with `contract.WriteResponse`.
- Mapped messaging, scheduler trigger, pubsub cluster publish, and webhook store/service failures to stable transport messages without exposing raw backend errors.
- Added envelope and sanitized-error coverage across messaging, pubsub, scheduler, and webhook tests.
- Validation passed:
  - `go test -race -timeout 60s ./x/messaging/... ./x/pubsub/... ./x/scheduler/... ./x/webhook/...`
  - `go test -timeout 20s ./x/messaging/... ./x/pubsub/... ./x/scheduler/... ./x/webhook/...`
  - `go vet ./x/messaging/... ./x/pubsub/... ./x/scheduler/... ./x/webhook/...`
