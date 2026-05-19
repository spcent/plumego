# Card 0118

Milestone:
Recipe: specs/change-recipes/stable-api.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files:
  - middleware/error_registry.go
  - middleware/concurrencylimit/concurrency_limit.go
  - middleware/coalesce/coalesce.go
  - middleware/timeout/timeout.go
  - middleware/recovery/recover.go
Depends On:

Goal:
Make the public middleware transport-error helper immutable. The stable package
currently exposes `WriteTransportError` as an exported variable, which lets
callers replace the canonical error path at runtime.

Scope:
- Convert `middleware.WriteTransportError` from an exported `var` to an exported function.
- Keep existing call sites source-compatible.
- Preserve the canonical `contract.WriteError` path and stable error codes.
- Add or adjust focused tests only if the conversion needs coverage.

Non-goals:
- Do not change error code values.
- Do not introduce new helper families.
- Do not change extension package error behavior beyond source compatibility.

Files:
- `middleware/error_registry.go`
- middleware call sites using `middleware.WriteTransportError`

Tests:
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Not required unless the public documentation describes `WriteTransportError` as assignable.

Done Definition:
- `WriteTransportError` is a function, not a variable.
- Middleware tests and dependency checks pass.

Outcome:
- Converted `middleware.WriteTransportError` from an exported assignable
  variable to an exported function that delegates to the internal transport
  helper.
- Preserved existing call syntax and stable middleware error code values.
- Validated with:
  - `go test -timeout 20s ./middleware/...`
  - `go vet ./middleware/...`
  - `go run ./internal/checks/dependency-rules`
