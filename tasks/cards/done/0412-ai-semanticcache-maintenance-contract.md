# Card 0412: AI Semantic Cache Maintenance Contract

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: done
Primary Module: x/ai
Owned Files:
- x/ai/semanticcache/cachemanager/manager.go
- x/ai/semanticcache/cachemanager/manager_test.go
- x/ai/semanticcache/cachemanager/warmer.go
- docs/modules/x-ai/README.md
Depends On: none

Goal:
- Replace vague semantic-cache maintenance stubs with an explicit capability contract.
- Make compaction, cleanup policy, and file warming failures predictable and test-covered.

Scope:
- Audit `CacheManager.Compact`, cleanup policy helpers, and `Warmer.WarmFromFile`.
- Replace generic `"not implemented"` errors with explicit sentinel errors or documented unsupported-operation behavior.
- Preserve existing cache get/set/delete semantics and vector store behavior.
- Add tests that assert the maintenance failure contract and supported no-op behavior where appropriate.

Non-goals:
- Do not implement a full persistence, compaction, or file import engine unless already locally supported.
- Do not change embedding, vector store, instrumentation, or cache hit semantics.
- Do not promote experimental semantic-cache APIs into stable provider/session contracts.

Files:
- `x/ai/semanticcache/cachemanager/manager.go`: define explicit maintenance unsupported behavior.
- `x/ai/semanticcache/cachemanager/manager_test.go`: cover compaction and cleanup policy contracts.
- `x/ai/semanticcache/cachemanager/warmer.go`: define explicit file warming unsupported behavior.
- `docs/modules/x-ai/README.md`: document maintenance capability expectations if public behavior changes.

Tests:
- `go test -race -timeout 60s ./x/ai/semanticcache/cachemanager`
- `go test -timeout 20s ./x/ai/semanticcache/cachemanager`
- `go vet ./x/ai/semanticcache/cachemanager`

Docs Sync:
- Required if public semantic-cache maintenance behavior or unsupported-operation errors change.

Done Definition:
- No semantic-cache maintenance method in scope returns a vague generic `"not implemented"` error.
- Unsupported maintenance behavior is represented by explicit, comparable errors or documented status.
- Focused tests cover each in-scope method.
- The three listed validation commands pass.

Outcome:
- Added comparable `ErrUnsupportedMaintenance` for cachemanager maintenance operations that require backend-specific capabilities.
- Replaced generic `not implemented` errors in `Compact`, selective cleanup policies, and `WarmFromFile`.
- Added focused `errors.Is` tests for compaction, oldest/least-used/expired cleanup, and file warming.
- Documented semantic-cache maintenance unsupported behavior in `docs/modules/x-ai/README.md`.
- Validation passed:
  - `go test -race -timeout 60s ./x/ai/semanticcache/cachemanager`
  - `go test -timeout 20s ./x/ai/semanticcache/cachemanager`
  - `go vet ./x/ai/semanticcache/cachemanager`
