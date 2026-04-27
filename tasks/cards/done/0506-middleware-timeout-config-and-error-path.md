# Card 0506

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files: middleware/timeout/timeout.go; middleware/timeout/timeout_test.go; tasks/cards/done/0506-middleware-timeout-config-and-error-path.md
Depends On: 2215

Goal:
Clarify timeout middleware defaults and keep all timeout/buffer failures on the canonical middleware transport error path.

Scope:
- Treat non-positive timeout duration as disabled pass-through middleware.
- Remove duplicated middleware imports and keep one package alias style.
- Replace the internal nil-request `contract.WriteError` bypass path with the canonical middleware transport error helper.
- Add tests for disabled timeout behavior and error envelope shape on replay/buffer failures where practical.

Non-goals:
- Do not redesign timeout streaming support.
- Do not add goroutine lifecycle abstractions beyond the existing handler wrapper.
- Do not change contract error builder behavior.

Files:
- `middleware/timeout/timeout.go`
- `middleware/timeout/timeout_test.go`

Tests:
- `go test -timeout 20s ./middleware/timeout`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; disabled non-positive timeouts are a defensive default correction for the existing constructor.

Done Definition:
- Zero timeout no longer causes immediate 504 responses.
- Timeout and buffer-failure responses use stable middleware error codes and canonical JSON shape.
- Targeted middleware tests and vet pass.

Outcome:
- Made non-positive timeout duration return pass-through middleware instead of creating an already-expired context.
- Removed the duplicate middleware import and kept a single alias.
- Switched timeout replay/bypass failure to `middleware.WriteTransportError`.
- Added disabled-timeout and error payload assertions.
- Validation run: `go test -timeout 20s ./middleware/timeout`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
