# Card 2139: x/pubsub Background Lifecycle Convergence

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/pubsub
Owned Files:
- `x/pubsub/pubsub.go`
- `x/pubsub/persistence.go`
- `x/pubsub/replay.go`
- `x/pubsub/ratelimit.go`
- `x/pubsub/pubsub_advanced_test.go`
Depends On: none

Goal:
Normalize background lifecycle ownership for `x/pubsub` wrappers so goroutines,
subscriptions, and cancellation paths are explicit and consistently closed.

Problem:
`x/pubsub` has many feature wrappers that create `context.WithCancel(context.Background())`
internally: persistence, replay, DLQ, rate limiting, ordering, audit,
distributed mode, consumer groups, multitenant mode, backpressure, and
schedulers. Most expose a `Close` method, but the construction and cancellation
pattern is not uniform, and new wrappers can easily forget idempotent close or
subscription cleanup. This is a lifecycle consistency risk in the largest
messaging subordinate package.

Scope:
- Establish a small internal lifecycle helper or documented local pattern for
  wrappers that own background goroutines.
- Apply the pattern to a representative high-risk slice: `PersistentPubSub`,
  `ReplayStore`, and `RateLimitedPubSub`.
- Add close-idempotency tests and at least one cancellation test that proves a
  background loop exits after `Close`.
- Document the required wrapper lifecycle pattern in `docs/modules/x-pubsub/README.md`
  if not already explicit.

Non-goals:
- Do not rewrite all pubsub feature wrappers in one pass.
- Do not change publish/subscribe semantics or message ordering.
- Do not introduce package-level global contexts.
- Do not move pubsub lifecycle ownership into `core` or stable middleware.

Files:
- `x/pubsub/pubsub.go`
- `x/pubsub/persistence.go`
- `x/pubsub/replay.go`
- `x/pubsub/ratelimit.go`
- `x/pubsub/pubsub_advanced_test.go`
- `docs/modules/x-pubsub/README.md`

Tests:
- `go test -race -timeout 60s ./x/pubsub/...`
- `go test -timeout 20s ./x/pubsub/...`
- `go vet ./x/pubsub/...`

Docs Sync:
Update `docs/modules/x-pubsub/README.md` with the background lifecycle rule if
the code pattern changes.

Done Definition:
- The selected pubsub wrappers use one clear background lifecycle pattern.
- `Close` is idempotent for the selected wrappers and covered by tests.
- At least one test proves background work stops after cancellation.
- The listed validation commands pass.

Outcome:
- Added package-local background lifecycle helpers for context creation and
  worker shutdown.
- Applied the helper to `PersistentPubSub`, `ReplayStore`, and
  `RateLimitedPubSub`.
- Made `ReplayStore.Close` idempotent and updated its existing close test.
- Added representative idempotent-close coverage and an adaptive-worker
  cancellation test.
- Documented the pubsub wrapper background lifecycle rule.

Validation:
- `go test -race -timeout 60s ./x/pubsub/...`
- `go test -timeout 20s ./x/pubsub/...`
- `go vet ./x/pubsub/...`
