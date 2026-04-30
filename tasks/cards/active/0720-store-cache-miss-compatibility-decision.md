# Card 0720

Milestone: —
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P2
State: active
Primary Module: store
Owned Files:
- `store/cache/cache.go`
- `store/cache/cache_test.go`
- `specs/deprecation-inventory.yaml`
- `docs/modules/store/README.md`
Depends On: —

Goal:
- Make an explicit compatibility decision for `store/cache.ErrCacheMiss`.

Problem:
`ErrCacheMiss` is retained as an alias for `ErrNotFound`, but the repository policy says deprecated symbols should not remain as dead wrappers once their callers are migrated. The alias may be a deliberate public compatibility exception, but that exception is not clearly recorded in the stable API/deprecation control plane.

Scope:
- Run the exported symbol completeness protocol for `ErrCacheMiss`.
- Choose one path:
  - formalize it as a supported compatibility alias in docs/control-plane data, or
  - remove it after all callers/tests are migrated to `ErrNotFound`.
- Keep cache miss semantics unchanged.
- Update tests to match the chosen policy.

Non-goals:
- Do not change cache storage behavior.
- Do not rename `ErrNotFound`.
- Do not alter unrelated store/cache errors.

Files:
- `store/cache/cache.go`
- `store/cache/cache_test.go`
- `specs/deprecation-inventory.yaml`
- `docs/modules/store/README.md`

Tests:
- `rg -n --glob '*.go' 'ErrCacheMiss' .`
- `go test -timeout 20s ./store/cache`
- `go run ./internal/checks/deprecation-inventory -strict`

Docs Sync:
- Required. Either document the compatibility alias or document its removal/migration path.

Done Definition:
- `ErrCacheMiss` has an explicit, reviewable policy decision.
- Search results are accounted for according to the symbol-change protocol.
- Deprecation inventory and tests agree with the chosen outcome.

Outcome:
-
