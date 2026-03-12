# Dependency Rules

Dependency rules exist to keep module ownership obvious.

## Global Rules

- Stable packages must not import `x/*`.
- Library packages must not import `reference/*`, `templates/*`, or `examples/*`.
- Root import `github.com/spcent/plumego` is forbidden as a long-term API shape.

## Stable Package Direction

- `contract` should remain closest to stdlib-only.
- `router` may depend on `contract`.
- `middleware` may depend on `contract`, `log`, `metrics`, and narrow `security` helpers.
- `security` may depend on `contract` and `log`.
- `store` may depend on `contract`, `log`, and `metrics`.
- `core` may depend on `router`, `middleware`, `contract`, `health`, `log`, and `metrics`.

## Extension Package Direction

- `x/*` packages may depend on stable packages.
- `x/tenant` owns tenant-aware middleware and tenant-aware store adapters.
- `x/ai` owns AI gateway behavior.
- `x/webhook`, `x/websocket`, `x/scheduler`, `x/frontend`, `x/messaging`, `x/discovery`, and `x/gateway` stay out of the stable path until explicitly stabilized.

## Migration Notes

While the repository still contains legacy roots such as `tenant`, `ai`, `net`, `pubsub`, `validator`, and `utils`, new work should follow the target layout instead of extending those roots further.
