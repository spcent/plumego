# Plan for M-016: Event-Driven Reference Architecture

Milestone: `M-016`
Objective: Ship reference/with-events demonstrating the outbox pattern, durable
consumer, retry loop, and webhook delivery using x/messaging primitives, with a
complete README that explains each pattern and how to adapt to a real broker.
Constraints: self-contained Go module importing only stable roots and x/messaging,
in-process pubsub backend only (no external broker), outbox pattern demonstrated
at application layer not framework layer, reference must run with `go run ./...`
and no required environment variables.
Affected Modules: reference/with-events.

## Phase Map

- Phase 1: Orient — read x/messaging sub-package APIs and standard-service
  canonical structure before scaffolding.
- Phase 2: Implement (parallel) — scaffold the app, implement order publisher
  and consumer, implement scheduler, implement webhook sender concurrently.
- Phase 3: Documentation — write README.md with ASCII diagram and run instructions.
- Phase 4: Validate and Ship — run acceptance criteria, commit.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1560 | Scaffold reference/with-events app skeleton | reference/with-events | `reference/with-events/main.go`, `reference/with-events/go.mod`, `reference/with-events/internal/config/config.go`, `reference/with-events/internal/app/app.go` | M-010 | `go build ./reference/with-events/...` |
| 1561 | Implement internal/order/ with outbox publisher and durable consumer | reference/with-events | `reference/with-events/internal/order/publisher.go`, `reference/with-events/internal/order/consumer.go`, `reference/with-events/internal/order/consumer_test.go` | 1560 | `go test ./reference/with-events/internal/order/...` |
| 1562 | Implement internal/scheduler/ with delayed retry job | reference/with-events | `reference/with-events/internal/scheduler/scheduler.go`, `reference/with-events/internal/scheduler/scheduler_test.go` | 1560 | `go test ./reference/with-events/internal/scheduler/...` |
| 1563 | Implement internal/webhook/ with WebhookSender and backoff retry | reference/with-events | `reference/with-events/internal/webhook/sender.go`, `reference/with-events/internal/webhook/sender_test.go` | 1560 | `go test ./reference/with-events/internal/webhook/...` |

## Dependency Edges

- `1560 -> 1561`
- `1560 -> 1562`
- `1560 -> 1563`

## Parallel Groups

- Group A: card 1560 — must complete first; provides the scaffold all others build on.
- Group B (parallel after A): cards 1561, 1562, 1563 — independent internal packages,
  no file overlap.
- Group C (sequential after B): README.md written after all four packages exist.

## Risk Register

- Risk: x/messaging in-process pubsub backend does not expose the interface needed
  for the outbox relay goroutine.
  Mitigation: card 1560 reads x/messaging/pubsub before scaffolding; if the interface
  is missing, record a blocker and stop — do not introduce workarounds in the reference.
- Risk: webhook retry loop blocks the test suite due to real HTTP calls.
  Mitigation: card 1563 uses net/http/httptest for the target URL; no real HTTP calls
  in any test.

## Verification Strategy

- Card-level checks: each implementation card runs its own `go test` immediately.
- Build check: `go build ./reference/with-events/...` after all four cards.
- Run check: `go run ./reference/with-events/...` starts without panicking and logs
  a startup message.
- Reference-layout check: `go run ./internal/checks/reference-layout` confirms the
  new reference follows canonical structure.

## Exit Condition

- all four implementation cards completed with tests
- reference/with-events builds and tests pass without external infrastructure
- README.md has ASCII diagram and run instructions
- verify report shows pass
- milestone acceptance criteria ready for PR packaging
