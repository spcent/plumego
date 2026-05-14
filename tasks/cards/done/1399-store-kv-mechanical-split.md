# Card 1399

Milestone: M-004
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/options.go
- store/kv/persistence.go
- store/kv/compat.go
- store/kv/kv_test.go
Depends On:
- 1394

Goal:
Mechanically split `store/kv` while preserving v1 compatibility methods and persistence behavior.

Scope:
- Move options/defaulting into a focused file.
- Move persistence/load/save helpers into a focused file.
- Move no-error compatibility methods such as `Exists`, `Keys`, `Size`, and `GetStats` into a focused file.
- Preserve context-aware methods as the preferred error-returning path.
- Keep all exported names and behavior unchanged.

Non-goals:
- Do not add durable KV-engine behavior to stable `store/kv`.
- Do not change file format, TTL cleanup, memory accounting, or close semantics.
- Do not remove compatibility methods.

Files:
- store/kv/kv.go
- store/kv/options.go
- store/kv/persistence.go
- store/kv/compat.go
- store/kv/kv_test.go

Tests:
- go test -race -timeout 60s ./store/kv
- go test -timeout 20s ./store/kv
- go vet ./store/kv

Docs Sync:
- Update `docs/modules/store/README.md` only if compatibility guidance is clarified.

Done Definition:
- `store/kv` exported API and persisted data behavior remain unchanged.
- Compatibility methods remain explicit and documented if docs are touched.
- Target checks pass.

Outcome:
- Mechanically split `store/kv/kv.go` into focused files for options,
  persistence, and no-error compatibility methods.
- Preserved exported API names, context-aware preferred methods, JSON state
  file format, and persistence behavior.
- Updated the deprecation inventory path for the moved compatibility comments.

Validation:
- go run ./internal/checks/deprecation-inventory -strict
- go test -race -timeout 60s ./store/kv
- go test -timeout 20s ./store/kv
- go vet ./store/kv
