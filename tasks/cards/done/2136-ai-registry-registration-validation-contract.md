# Card 2136: AI Registry Registration Validation Contract
Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: active
Primary Module: x/ai
Owned Files:
- x/ai/provider/manager.go
- x/ai/provider/provider_test.go
- x/ai/streaming/streaming.go
- x/ai/streaming/streaming_test.go
- docs/modules/x-ai/README.md
Depends On: none

Goal:
Add explicit registration validation paths for small x/ai registries. Provider
manager registration currently dereferences `provider.Name()` without a nil
guard, and streaming registration accepts empty workflow IDs or nil streams.
Both are easy-to-miss runtime input errors that should fail at registration time
through a documented path.

Scope:
- Add error-returning registration methods for provider and streaming managers.
- Preserve existing `Register` methods as compatibility wrappers if needed.
- Add tests for nil provider, empty provider name, empty workflow ID, nil stream,
  and valid registration behavior.
- Keep manager routing and SSE stream delivery behavior unchanged.

Non-goals:
- Do not redesign provider routing.
- Do not change SSE event format or stream lifecycle.
- Do not add global registries or package-level registration hooks.

Files:
- `x/ai/provider/manager.go`
- `x/ai/provider/provider_test.go`
- `x/ai/streaming/streaming.go`
- `x/ai/streaming/streaming_test.go`
- `docs/modules/x-ai/README.md`

Tests:
- go test -race -timeout 60s ./x/ai/provider ./x/ai/streaming
- go test -timeout 20s ./x/ai/provider ./x/ai/streaming
- go vet ./x/ai/provider ./x/ai/streaming

Docs Sync:
Update `docs/modules/x-ai/README.md` only if it documents provider or streaming
registration.

Done Definition:
- Runtime registration input errors have explicit error-returning paths.
- Existing valid `Register` callers remain compatible.
- Provider routing and streaming delivery tests still pass.
- The listed validation commands pass.

Outcome:
- Added explicit error-returning `RegisterE` methods for provider manager and
  streaming stream manager registration.
- Added provider and streaming registration sentinels for nil provider, empty
  provider name, empty workflow ID, and nil SSE stream.
- Preserved existing `Register` methods as compatibility wrappers for
  known-good inputs, with invalid input failing at registration time.
- Migrated `ExecuteStreaming` to use the explicit stream registration error
  path.
- Updated x-ai module documentation to point dynamic registration callers at
  `RegisterE`.
- Validation passed:
  - `go test -race -timeout 60s ./x/ai/provider ./x/ai/streaming`
  - `go test -timeout 20s ./x/ai/provider ./x/ai/streaming`
  - `go vet ./x/ai/provider ./x/ai/streaming`
