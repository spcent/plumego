# Card 1416

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/ai
Owned Files:
- x/ai/provider/manager.go
- x/ai/provider/provider_test.go
- x/ai/streaming/streaming.go
- x/ai/streaming/streaming_test.go
- docs/modules/x-ai/README.md
Depends On:
- 1415

Goal:
- Prefer error-returning dynamic registration paths in stable-tier `x/ai` subpackages while retaining compatibility wrappers.

Scope:
- Review `provider` and `streaming` panic-style dynamic registration helpers.
- Ensure error-returning alternatives are documented and covered by tests.
- Keep compatibility wrappers for known-good internal wiring.
- Avoid touching experimental AI orchestration/cache packages in this card.

Non-goals:
- Do not change provider request/response contracts.
- Do not change streaming event or SSE payload shapes.
- Do not migrate AI resilience types.

Files:
- x/ai/provider/manager.go
- x/ai/provider/provider_test.go
- x/ai/streaming/streaming.go
- x/ai/streaming/streaming_test.go
- docs/modules/x-ai/README.md

Tests:
- go test -timeout 20s ./x/ai/provider ./x/ai/streaming
- go vet ./x/ai/provider ./x/ai/streaming
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update `docs/modules/x-ai/README.md` with constructor and dynamic registration guidance.

Done Definition:
- Stable-tier dynamic registration has explicit error-returning guidance.
- Compatibility wrappers remain behavior-compatible.
- Provider and streaming tests pass.

Outcome:
- Completed on 2026-05-15.
- Clarified `provider.Manager.Register` and `streaming.StreamManager.Register` as compatibility wrappers for known-good wiring.
- Preserved `RegisterE` as the explicit error-returning dynamic registration path.
- Tightened provider and streaming tests to assert compatibility wrappers panic with the underlying `RegisterE` validation errors.
- Updated `docs/modules/x-ai/README.md` with dynamic registration guidance.
- Validation:
  - `go test -timeout 20s ./x/ai/provider ./x/ai/streaming`
  - `go vet ./x/ai/provider ./x/ai/streaming`
  - `go run ./internal/checks/dependency-rules`
