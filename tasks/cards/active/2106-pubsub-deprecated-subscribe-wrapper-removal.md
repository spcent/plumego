# Card 2106: Remove Deprecated PubSub Subscribe Wrapper

Priority: P1
State: active
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: x/pubsub
Owned Files:
- x/pubsub/pubsub.go
- x/pubsub/pubsub_advanced_test.go
- x/pubsub/pubsub_basic_test.go
- docs/modules/x-pubsub/README.md
Depends On:

## Goal

`x/pubsub/pubsub.go` still exports `SubscribeWithContext` with a deprecation
comment:

```go
// Deprecated: Use Subscribe(ctx, topic, opts) instead.
func (b *InProcBroker) SubscribeWithContext(...)
```

AGENTS.md requires deprecated symbols to be removed in the same PR that
replaces their last caller.  A current repo search shows no production caller
besides the definition, so the wrapper should be removed rather than kept as a
dead compatibility path.

## Scope

- Follow the exported-symbol change protocol before editing.
- Remove `SubscribeWithContext` if no callers remain.
- Update tests or docs that mention the deprecated name.
- Keep `Subscribe(ctx, topic, opts)` as the canonical exact-topic subscription
  API.

## Non-goals

- Do not rename `SubscribePatternWithContext` or `SubscribeMQTTWithContext` in
  this card; those are separate APIs that still define a context-bearing
  variant for APIs whose base method does not take context.
- Do not change subscription delivery, cancellation, ring-buffer, or MQTT
  behavior.
- Do not add compatibility wrappers.

## Files

- `x/pubsub/pubsub.go`
- `x/pubsub/pubsub_advanced_test.go`
- `x/pubsub/pubsub_basic_test.go`
- `docs/modules/x-pubsub/README.md`

## Tests

```bash
rg -n --glob '*.go' 'SubscribeWithContext' .
go test -race -timeout 60s ./x/pubsub/...
go vet ./x/pubsub/...
```

## Docs Sync

Update `docs/modules/x-pubsub/README.md` only if it mentions the deprecated
symbol or needs to call out the canonical `Subscribe(ctx, topic, opts)` shape.

## Done Definition

- The pre-edit `rg` result is documented in the implementation summary.
- No `SubscribeWithContext` references remain after editing.
- Tests still cover context-driven cancellation through `Subscribe`.
- The listed validation commands pass.

## Outcome

