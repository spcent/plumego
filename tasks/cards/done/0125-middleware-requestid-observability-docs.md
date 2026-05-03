# Card 0125

Milestone:
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
  - docs/modules/middleware/README.md
  - middleware/requestid/generator.go
  - middleware/requestid/request_id.go
  - middleware/accesslog/accesslog.go
  - middleware/module.yaml
Depends On: 0124

Goal:
Clarify request-id generation and stable observability ownership so callers can
wire deterministic production behavior without relying on hidden assumptions.

Scope:
- Document that `requestid.WithGenerator` is the explicit production override and `NewRequestID` uses package-local generator state.
- Clarify that `accesslog` observer/tracer parameters are compatibility wiring for transport observability, not exporter or backend ownership.
- Add concise Go doc where public symbols are under-documented.

Non-goals:
- Do not remove or rename accesslog parameters.
- Do not migrate tracing to `x/observability` in this card.
- Do not change runtime behavior.

Files:
- `docs/modules/middleware/README.md`
- `middleware/requestid/generator.go`
- `middleware/requestid/request_id.go`
- `middleware/accesslog/accesslog.go`
- `middleware/module.yaml`

Tests:
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- `docs/modules/middleware/README.md`
- `middleware/module.yaml`

Done Definition:
- Public request-id generation behavior is documented.
- Stable observability ownership boundaries are documented.
- Middleware tests and manifest checks pass.

Outcome:
- Added Go docs for request-id options and middleware, including the
  `WithGenerator` production override path.
- Clarified `NewRequestID` uses package-local default generator state.
- Documented `accesslog` observer/tracer arguments as transport-only wiring,
  with exporter/backend ownership outside stable middleware.
- Synced the same ownership guidance into `docs/modules/middleware/README.md`
  and `middleware/module.yaml`.
- Validated with:
  - `go test -timeout 20s ./middleware/...`
  - `go vet ./middleware/...`
  - `go run ./internal/checks/module-manifests`
