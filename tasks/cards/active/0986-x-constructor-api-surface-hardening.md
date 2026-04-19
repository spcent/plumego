# Card 0986: x Package Constructor API Surface Hardening

Priority: P2
State: active
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/gateway, x/mq

## Goal

Two x packages have constructor or option-func APIs that panic where errors would be more
appropriate, with no safe alternative provided.  This contrasts with the established
convention in `x/data/sharding`, `x/data/rw`, and `x/websocket`, which return `(*T, error)`.

### Issue 1 — `x/gateway.New()` panics, no `NewE` variant

`x/gateway/proxy.go:64`:

```go
func New(config Config) *Proxy {
    cfg := config.WithDefaults()
    if err := cfg.Validate(); err != nil {
        panic(err)                   // line 70
    }
    pool, err = NewBackendPool(cfg.Targets)
    if err != nil {
        panic(err)                   // line 80
    }
    ...
}
```

`x/mq` and `x/pubsub` both provide `NewInProcBrokerE` / `NewInProcBroker` pairs.  Gateway
has no safe constructor.  Callers that want to handle misconfiguration without a panic (e.g.
integration tests, dynamic config) have no recourse.

**Fix:** Add `NewE(config Config) (*Proxy, error)` that returns the error instead of panicking.
Keep `New()` as a convenience that calls `NewE` and panics on error (document it).

### Issue 2 — `x/mq.WithConfig()` option func panics internally

`x/mq/broker.go:107-113`:

```go
func WithConfig(cfg Config) Option {
    return func(b *InProcBroker) {
        if err := cfg.Validate(); err != nil {
            panic(fmt.Sprintf("invalid broker config: %v", err))  // line 110
        }
        b.config = cfg
    }
}
```

Functional options must not panic.  Callers compose options before construction and have no
way to recover from a panic inside an option application.  The correct place for validation is
`newInProcBroker`, not inside the option closure.

**Fix:** Remove the `Validate()` call from `WithConfig`.  Unconditionally assign `b.config = cfg`.
`newInProcBroker` already calls `cfg.Validate()` as part of construction, so the validation
still runs — at the right time, with an error return path.

## Scope

- `x/gateway/proxy.go`: add `NewE(config Config) (*Proxy, error)`, refactor `New` to call it.
- `x/mq/broker.go`: remove early `cfg.Validate()` from `WithConfig` closure.
- Update `x/gateway/module.yaml` agent_hints to document the `NewE`/`New` pair.

## Non-goals

- Do not change other gateway constructors (backend pool, transport pool, health checker).
- Do not refactor `x/mq.NewInProcBroker` / `NewInProcBrokerE` — they are already correct.
- Do not rename or remove `New()` in x/gateway — breaking change, keep it as the panic
  convenience wrapper.

## Files

- `x/gateway/proxy.go`
- `x/mq/broker.go` (WithConfig function only)
- `x/gateway/module.yaml` (agent_hints)

## Tests

```bash
go test -timeout 20s ./x/gateway/...
go test -timeout 20s ./x/mq/...
go vet ./x/gateway/... ./x/mq/...
```

Verify that `WithConfig` with an invalid config no longer panics but that construction still
fails with an error:

```go
// In a new test:
b, err := mq.NewInProcBrokerE(ps, mq.WithConfig(invalidCfg))
// err != nil, b == nil — no panic
```

## Done Definition

- `x/gateway.NewE(config Config) (*Proxy, error)` exists and returns a non-nil error for
  invalid config or backend pool failure instead of panicking.
- `x/gateway.New` delegates to `NewE` and panics on error (documented).
- `x/mq.WithConfig` option closure does not call `cfg.Validate()`.
- `go test ./x/gateway/... ./x/mq/...` passes.
- `go vet` clean.
