# Card 0718

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/messaging
Owned Files:
- `x/messaging/service.go`
- `x/messaging/runtime.go`
- `x/messaging/service_test.go`
- `docs/modules/x-messaging/README.md`
Depends On: —

Goal:
- Reduce `messaging.Service` responsibility by extracting worker runtime defaults and queue runtime assembly out of the central service constructor.

Problem:
`Service` currently owns queue, worker, store, scheduler, pubsub, templates, providers, receipts, quota, monitor, metrics, webhook notifier, and counters. `New` also silently creates an in-memory task store and hard-coded worker defaults. This makes the app-facing service hard to review and hard to evolve safely.

Scope:
- Introduce a small internal runtime assembly type or helper in `x/messaging/runtime.go`.
- Move worker defaulting and `mq.NewTaskQueue` / `mq.NewWorker` construction behind that helper.
- Keep the public `Config` and `New(Config) *Service` API stable.
- Add or update tests that assert default task store, worker registration, and explicit worker tuning still work.
- Document the new internal ownership split in the module README if it helps future work.

Non-goals:
- Do not split provider delivery, scheduled maintenance, or webhook notification in this card.
- Do not change public `Send`, `Start`, `Stop`, `Stats`, or receipt behavior.
- Do not remove the default in-memory store yet.

Files:
- `x/messaging/service.go`
- `x/messaging/runtime.go`
- `x/messaging/service_test.go`
- `docs/modules/x-messaging/README.md`

Tests:
- `go test -timeout 20s ./x/messaging/...`
- `go vet ./x/messaging/...`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Required if ownership split is described in docs.

Done Definition:
- Worker runtime defaulting is no longer embedded directly in `Service.New`.
- Public messaging API remains compatible.
- Tests confirm existing send/worker behavior still passes.

Outcome:
-
