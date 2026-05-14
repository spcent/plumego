# Card 1410

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/mq
Owned Files:
- x/mq/broker.go
- x/mq/broker_runtime.go
- x/mq/broker_routes.go
- x/mq/mq_test.go
- x/mq/queue_worker_test.go
Depends On:
- 1409

Goal:
- Split `x/mq` broker runtime responsibilities while preserving explicit unsupported bridge behavior.

Scope:
- Move broker loop, handler dispatch, and worker coordination helpers into `broker_runtime.go`.
- Move broker-facing registration or route-like helpers into `broker_routes.go` where applicable.
- Preserve MQTT/AMQP unsupported operation behavior and config validation.

Non-goals:
- Do not implement MQTT or AMQP.
- Do not change ack defaults, trie defaults, or queue persistence semantics.
- Do not change app-facing `x/messaging` behavior.

Files:
- x/mq/broker.go
- x/mq/broker_runtime.go
- x/mq/broker_routes.go
- x/mq/mq_test.go
- x/mq/queue_worker_test.go

Tests:
- go test -timeout 30s ./x/mq/...
- go vet ./x/mq/...
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update `docs/modules/x-mq/README.md` only if unsupported bridge wording changes.

Done Definition:
- Broker runtime has narrower file ownership.
- Unsupported bridge tests still pass.
- No queue, ack, or persistence behavior changes are introduced.

Outcome:
- Completed on May 15, 2026.
- Split memory sampling, observability wrapping, shutdown, panic handling, and metrics observation helpers into `x/mq/broker_runtime.go`.
- Split cluster bridge helpers, explicit MQTT/AMQP unsupported entrypoints, persistence replay helpers, and consumer-group accessor into `x/mq/broker_routes.go`.
- Kept `x/mq/broker.go` focused on broker construction, validation, and base publish/subscribe behavior.
- Preserved unsupported MQTT/AMQP behavior and queue, ack, persistence, and transaction semantics.
- Validation passed:
  - `go test -timeout 30s ./x/mq/...`
  - `go vet ./x/mq/...`
  - `go run ./internal/checks/dependency-rules`
