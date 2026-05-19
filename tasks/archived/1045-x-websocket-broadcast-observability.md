# Card 1045

Milestone:
Recipe: specs/change-recipes/api-cleanup.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`
Depends On: 0743

Goal:
- Make broadcast fanout outcomes observable and actionable for callers.

Problem:
Dropped broadcast messages are counted only when `EnableMetrics` is true, even though drops are runtime facts. `RejectOnQueueFull` also cannot be observed by callers of `BroadcastRoom` or `BroadcastAll`, so applications cannot react to queue saturation.

Scope:
- Track broadcast attempted, delivered, skipped, and dropped counts independently of the metrics toggle.
- Add explicit result-returning broadcast APIs such as `TryBroadcastRoom` and `TryBroadcastAll`, or change existing APIs to return a `BroadcastResult`.
- Update the admin broadcast endpoint to use the result-returning path and expose appropriate HTTP errors for total rejection.
- Keep metrics export as a view over runtime counters, not the source of truth.
- Update tests for partial delivery, total rejection, metrics disabled, and metrics enabled.

Non-goals:
- Do not add persistent delivery guarantees.
- Do not introduce per-message acknowledgements.
- Do not turn Hub into a generic queue broker.

Files:
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required for broadcast return values, drop counters, and admin endpoint responses.

Done Definition:
- Queue-full drops are counted even when metrics are disabled.
- Callers can distinguish full success, partial delivery, and total rejection.
- Admin broadcast no longer reports success when no connection could accept the message.
- Metrics docs match the final counter source and names.

Outcome:
- Added `BroadcastResult`, `TryBroadcastRoom`, and `TryBroadcastAll`.
- Kept `BroadcastRoom` and `BroadcastAll` as result-ignoring convenience methods.
- Tracked attempted, enqueued, skipped, and dropped broadcast counters independently of `EnableMetrics`.
- Updated admin broadcast to return `WEBSOCKET_BROADCAST_REJECTED` on total queue rejection.
- Added tests for partial delivery, total rejection, metrics-disabled drop accounting, and admin rejection.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go test -race -timeout 60s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go build ./...`
