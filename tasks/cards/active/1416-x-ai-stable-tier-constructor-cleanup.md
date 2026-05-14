# Card 1416

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
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

