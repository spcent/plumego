# Card 0325

Milestone:
Recipe: specs/change-recipes/refine-docs.yaml
Priority: P2
State: active
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/db/sql.go
Depends On:
- 0324-store-kv-read-lock-convergence

Goal:
Polish package-level examples so stable `store` docs are complete and copyable.

Scope:
- Add missing imports and placeholder values to the `store/cache` package example.
- Add the missing SQL import to the `store/db` package example.
- Keep examples scoped to stable primitives and caller-owned context/deadlines.

Non-goals:
- Do not add runnable integration examples requiring external services.
- Do not change runtime behavior.
- Do not edit extension package docs.

Files:
- store/cache/cache.go
- store/db/sql.go

Tests:
- go test -timeout 20s ./store/cache ./store/db
- go vet ./store/cache ./store/db

Docs Sync:
- Not required; package comments only.

Done Definition:
- Package examples no longer reference missing imports or undeclared values.
- Examples remain stable-layer scoped.
- Targeted tests and vet pass.
