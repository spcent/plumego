# Card 0012

Priority: P2
State: done
Primary Module: x/pubsub
Owned Files:
  - x/pubsub/pubsub.go
  - x/pubsub/ack.go

Depends On: —

Goal:
`x/pubsub.InProcBroker`'s public API is severely inconsistent with respect to context handling:

**Publish side**:
- `Publish(topic, msg)` — no context
- `PublishWithContext(ctx, topic, msg)` — has context (two variants of the same operation coexist)
- `PublishAsync(topic, msg)` — no context
- `PublishBatch(topic, msgs)` — no context

**Subscribe side**:
- `Subscribe(topic, opts)` — no context
- `SubscribeAckable(topic, subOpts, ackOpts)` — no context
- `SubscribeOrdered(ctx, topic, opts)` — has context

API consumers cannot form a consistent mental model: when is ctx needed for subscribing? Which publish variant should be used?
Having `Publish` + `PublishWithContext` coexist is obvious API surface noise.

Scope:
- **Publish-side unification**: change `Publish(topic, msg)` to the preferred `Publish(ctx, topic, msg)`,
  make the internal `PublishWithContext` private or delete it (retain as a deprecated wrapper during
  the backward-compatibility transition period);
  `PublishAsync` and `PublishBatch` similarly gain a ctx parameter
- **Subscribe-side ctx addition**: add a ctx parameter to `Subscribe(topic, opts)` and
  `SubscribeAckable(topic, subOpts, ackOpts)`, used to control subscription lifecycle (ctx cancel
  triggers subscription close, consistent with SubscribeOrdered)
- Update all internal callers and tests
- If the change scope is too large (more than 20 call sites), split into two steps: this card first
  completes Subscribe/SubscribeAckable ctx addition; publish side gets a separate card

Non-goals:
- Do not change message delivery semantics (at-most-once / at-least-once)
- Do not modify Message, SubOptions, AckOptions, or other data structures
- Do not affect the behavior of RateLimitedPubSub / DistributedPubSub wrappers
  (just follow the main type's signature changes)

Files:
  - x/pubsub/pubsub.go (Publish family signatures)
  - x/pubsub/ack.go (add ctx to SubscribeAckable)
  - x/pubsub/ratelimit.go (follow-along adjustment)
  - x/pubsub/distributed.go (follow-along adjustment)
  - x/pubsub/*_test.go (update callers)

Tests:
  - go test -race ./x/pubsub/...

Docs Sync: —

Done Definition:
- `Subscribe` and `SubscribeAckable` both accept `ctx context.Context` as the first parameter
- `Publish` and `PublishWithContext` resolved to one choice (preferred: Publish with ctx; old variant
  marked Deprecated or deleted)
- `go test -race ./x/pubsub/...` passes
- `grep -n "PublishWithContext" x/pubsub/*.go` returns empty or only the deprecated wrapper

Outcome:
- Changed `Subscribe` signature to `Subscribe(ctx context.Context, topic string, opts SubOptions) (Subscription, error)`
- Added `SubscribeWithContext` as a deprecated alias calling the new signature
- Updated `Broker` interface in broker.go
- Updated `SubscribeAckable` in ack.go similarly
- Fixed all callers: x/mq/ack.go, x/mq/broker.go, x/webhook/bridge.go, x/messaging/service_test.go, test mocks in mq_test.go and priority_test.go
