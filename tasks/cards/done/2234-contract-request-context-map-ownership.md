# Card 2234

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/context_core.go
- contract/context_test.go
- contract/context_extended_test.go
Depends On: 2233

Goal:
Make request metadata maps owned by `contract` carriers instead of aliasing caller-owned maps.

Scope:
- Copy `RequestContext.Params` when storing and when returning from context.
- Copy `params` passed into `NewCtx` / `NewCtxWithConfig`.
- Add focused tests proving caller mutation and returned-map mutation do not leak back into stored context values.

Non-goals:
- Do not make `Ctx.Params` private.
- Do not add request bags or string-key service locator behavior.
- Do not change route param lookup semantics.

Files:
- `contract/context_core.go`
- `contract/context_test.go`
- `contract/context_extended_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this hardens existing request metadata behavior.

Done Definition:
- `WithRequestContext` and `RequestContextFromContext` use defensive map copies.
- `NewCtx` does not alias the caller's params map.
- Existing context and binding tests continue to pass.

Outcome:
- Added defensive copying for request context params on store and lookup.
- Made `NewCtx` / `NewCtxWithConfig` own a copy of route params passed by callers.
- Validation run: `go test -timeout 20s ./contract/...`; `go vet ./contract/...`.
