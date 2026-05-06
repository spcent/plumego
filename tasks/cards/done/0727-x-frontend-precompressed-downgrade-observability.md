# Card 0727

Milestone:
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files: x/frontend/config.go, x/frontend/compression.go, x/frontend/compression_test.go, x/frontend/README.md, docs/modules/x-frontend/README.md, docs/extension-evidence/x-frontend.md
Depends On: 0726

Goal:
Provide an explicit application-owned observability hook for precompressed variant downgrade events.

Scope:
- Add a small exported event type and `WithPrecompressedVariantMissHandler` option.
- Emit the hook when an accepted `.br` or `.gz` candidate cannot be opened or statted and x/frontend downgrades or continues probing.
- Keep the default behavior unchanged: no logging, metrics, globals, or extra dependency when the hook is not configured.
- Add tests for stat/open miss reporting without changing response behavior.
- Update docs and evidence to remove the “no signal exists” stable decision blocker.

Non-goals:
- Do not add built-in logging or metrics.
- Do not expose raw filesystem errors that may leak local paths.
- Do not change identity fallback or 406 semantics.

Files:
- `x/frontend/config.go`
- `x/frontend/compression.go`
- `x/frontend/compression_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
- Document the hook as application-owned observability for stale or damaged build artifacts.

Done Definition:
- Applications can observe precompressed variant misses through an explicit option.
- Existing default response behavior is unchanged.
- Targeted x/frontend tests and vet pass.

Outcome:
- Added `PrecompressedVariantMiss` and `WithPrecompressedVariantMissHandler`.
- Emitted application-owned miss events for planned variant open misses and accepted variant stat misses.
- Documented the hook in package docs, module docs, and the x/frontend evidence ledger.
- Validation Run:
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
