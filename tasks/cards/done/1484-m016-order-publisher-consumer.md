# Card 1561

Milestone: M-016
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P2
State: done
Primary Module: reference/with-events
Owned Files:
- `reference/with-events/internal/order/publisher.go`
- `reference/with-events/internal/order/consumer.go`
- `reference/with-events/internal/order/handler.go`
- `reference/with-events/internal/order/handler_test.go`

Goal:
- Implement the order event flow in reference/with-events: an HTTP handler
  writes an OrderCreated event to the pubsub topic (outbox pattern), and a
  consumer goroutine processes it idempotently using x/messaging/pubsub.

Scope:
- Create internal/order/publisher.go:
  - OrderPublisher struct holding a pubsub.Publisher and an idempotency store.
  - Publish(ctx, OrderCreated) error — publishes to "orders.created" topic;
    records message ID in idempotency store before publishing.
- Create internal/order/consumer.go:
  - OrderConsumer struct holding a pubsub.Subscriber and idempotency store.
  - Start(ctx context.Context) — subscribes to "orders.created"; for each
    message checks idempotency store (skip if already processed), processes,
    marks done, acks; uses at-least-once semantics.
- Create internal/order/handler.go:
  - POST /orders handler: decodes CreateOrderRequest, calls publisher.Publish,
    returns 202 Accepted via contract.WriteResponse.
  - GET /orders/:id handler: returns stub order from in-memory store.
- Write internal/order/handler_test.go covering:
  - POST /orders with valid body returns 202.
  - POST /orders with empty body returns 400.
  - Consumer skips duplicate message (idempotency check).
  - Consumer processes new message and marks it done.

Non-goals:
- Do not use a real message broker; use the in-process pubsub backend.
- Do not implement order persistence beyond an in-memory map.
- Do not add payment or fulfilment logic.

Files:
- `reference/with-events/internal/order/publisher.go`
- `reference/with-events/internal/order/consumer.go`
- `reference/with-events/internal/order/handler.go`
- `reference/with-events/internal/order/handler_test.go`

Tests:
- `go test -timeout 30s ./reference/with-events/internal/order/...`
- `go vet ./reference/with-events/...`
- `go build ./reference/with-events/...`

Docs Sync:
- none at this card; README.md written in M-016 Phase 3.

Done Definition:
- POST /orders returns 202 for valid input.
- Consumer skips duplicate events; processes new ones.
- All four handler_test.go cases pass.
- `go build ./reference/with-events/...` exits 0.

Outcome:
- Added the `internal/order` package with `OrderPublisher`,
  `OrderConsumer`, in-memory idempotency store, HTTP handler, and tests.
- Wired `POST /orders` and `GET /orders/:id` into the with-events app,
  replacing the scaffold placeholders for the order route group.
- The publisher records event IDs before publishing to `orders.created`; the
  consumer skips duplicate message IDs and marks new messages as processed.
- Validation passed with order package tests, `reference/with-events` build,
  `reference/with-events` vet, reference-layout, agent-workflow, and
  `git diff --check`.
