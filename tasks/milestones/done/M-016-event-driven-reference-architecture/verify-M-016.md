# Verify M-016: Event-Driven Reference Architecture

Milestone: `M-016`
Branch: implemented on main (reference/with-events landed before milestone tracking was closed)
Verified: 2026-05-26

## Scope Check

- In-scope files touched: `reference/with-events/` (main.go, go.mod, internal/app/, internal/config/, internal/order/, internal/scheduler/, internal/webhook/, README.md)
- Out-of-scope files touched: none
- x/messaging packages unmodified: confirmed

## Implementation Completeness

| Component | Status |
|---|---|
| `internal/app/` — core.App, middleware chain, messaging service wiring | DONE |
| `internal/config/` — config load, defaults, validation | DONE |
| `internal/order/` — OrderPublisher (outbox + pubsub relay) and OrderConsumer (idempotent) | DONE |
| `internal/scheduler/` — delayed retry job via x/messaging/scheduler | DONE |
| `internal/webhook/` — WebhookSender with configurable target URL and backoff retry | DONE |
| `README.md` — ASCII architecture diagram, run instructions, pattern explanations | DONE |

## Acceptance Test Results

| Check | Result |
|---|---|
| `go build ./reference/with-events/...` | PASS |
| `go test -timeout 30s ./reference/with-events/...` | PASS (order, scheduler, webhook suites) |
| `go vet ./reference/with-events/...` | PASS |
| `go run ./internal/checks/dependency-rules` | PASS |
| `go run ./internal/checks/reference-layout` | PASS |
| `go run ./internal/checks/module-manifests` | PASS |
| `go run ./internal/checks/agent-workflow` | PASS |
| No external broker or infrastructure required | CONFIRMED (in-process pubsub backend) |

## Deviations from Spec

- Phase 1 step 2 ("confirm with-events does not exist") was moot — the reference was implemented directly. Milestone tracking was not updated as tasks completed.
- No separate branch was pushed; work landed via the normal PR flow.
- No separate PR with the title `milestone(M-016): Event-Driven Reference Architecture` was opened; the work was included in a broader PR.

These are process deviations only. All functional acceptance criteria pass.
