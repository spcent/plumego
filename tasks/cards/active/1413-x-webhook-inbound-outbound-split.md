# Card 1413

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/webhook
Owned Files:
- x/webhook/outbound_service.go
- x/webhook/outbound_dispatch.go
- x/webhook/inbound_hmac.go
- x/webhook/inbound_provider_errors.go
- x/webhook/outbound_test.go
Depends On:
- 1412

Goal:
- Split webhook outbound dispatch and inbound provider error mapping without changing security semantics.

Scope:
- Move outbound dispatch/retry helpers into `outbound_dispatch.go`.
- Move provider-specific inbound error mapping helpers into `inbound_provider_errors.go`.
- Preserve HMAC verification, provider mapping, retry, and failure behavior.

Non-goals:
- Do not change signature verification algorithms.
- Do not add new webhook providers.
- Do not change outbound delivery semantics.

Files:
- x/webhook/outbound_service.go
- x/webhook/outbound_dispatch.go
- x/webhook/inbound_hmac.go
- x/webhook/inbound_provider_errors.go
- x/webhook/outbound_test.go

Tests:
- go test -timeout 30s ./x/webhook
- go vet ./x/webhook
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update docs only if provider error behavior is documented differently.

Done Definition:
- Outbound dispatch and inbound error mapping have separate file ownership.
- Existing webhook security and outbound tests pass.
- No provider or signature behavior change is introduced.

Outcome:

