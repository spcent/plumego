# Card 2129: AI Semantic Cache Provider Constructor Contract
Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: done
Primary Module: x/ai
Owned Files:
- x/ai/semanticcache/provider.go
- x/ai/semanticcache/provider_test.go
- docs/modules/x-ai/README.md
Depends On: none

Goal:
Make semantic-cache provider construction explicit about required dependencies.
`NewSemanticCachingProvider` currently accepts a provider and cache pointer
without validation; nil values produce later method panics in `Name`, `Complete`,
or cache access paths.

Scope:
- Add an error-returning constructor for `SemanticCachingProvider` that validates
  required dependencies.
- Preserve the existing constructor as a compatibility wrapper if needed, but
  make invalid inputs fail at construction time rather than through unrelated
  later nil dereferences.
- Add focused tests for nil provider, nil cache, and valid passthrough behavior.
- Keep provider/cache ownership in `x/ai/semanticcache`; do not move lifecycle
  management into stable roots.

Non-goals:
- Do not redesign semantic-cache storage or similarity behavior.
- Do not add logging infrastructure to hide cache store errors.
- Do not change exported provider interfaces unless required for validation.

Tests:
- go test -race -timeout 60s ./x/ai/semanticcache
- go test -timeout 20s ./x/ai/semanticcache
- go vet ./x/ai/semanticcache

Docs Sync:
Update `docs/modules/x-ai/README.md` if it documents semantic-cache provider
construction.

Done Definition:
- Invalid semantic-cache provider construction has a documented, test-covered
  error path.
- Existing valid constructor behavior remains compatible.
- The semanticcache validation commands pass.
- No stable root depends on x/ai.

Outcome:
Added `NewSemanticCachingProviderE` with explicit sentinel errors for missing
upstream provider and missing semantic cache. The existing
`NewSemanticCachingProvider` signature is preserved as a compatibility wrapper
that panics immediately on invalid dependencies. Tests now cover valid
construction, nil provider, nil cache, nil config fallback, and compatibility
panic behavior.

Validation:
- `go test -race -timeout 60s ./x/ai/semanticcache`
- `go test -timeout 20s ./x/ai/semanticcache`
- `go vet ./x/ai/semanticcache`
