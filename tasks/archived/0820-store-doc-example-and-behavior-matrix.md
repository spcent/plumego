# Card 0820: Store Doc Example And Behavior Matrix

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P0
State: done
Primary Module: store
Owned Files:
- store/idempotency/store.go
- docs/modules/store/README.md
- store/file/README.md
- store/file/TESTING.md
Depends On:

Goal:
Fix stale store documentation examples and make the stable store subpackage behavior matrix explicit before further API work.

Scope:
- Correct the `store/idempotency` package example for the current `x/data/idempotency.NewSQLStore` signature.
- Add a concise behavior matrix covering missing, expired, closed, invalid key, nil context, and delete-missing semantics across `cache`, `kv`, `file`, `idempotency`, and `db`.
- Sync `store/file` README/testing wording with the current contract comments.

Non-goals:
- Do not change code behavior in this card.
- Do not add provider-specific file behavior to stable `store/file`.
- Do not refresh the stable API snapshot in this card.

Files:
- store/idempotency/store.go
- docs/modules/store/README.md
- store/file/README.md
- store/file/TESTING.md

Tests:
- go test -timeout 20s ./store/...
- go vet ./store/...
- go run ./internal/checks/module-manifests

Docs Sync:
- Required.

Done Definition:
- The idempotency example compiles conceptually against the current constructor signatures.
- Store behavior matrix documents the intended stable semantics.
- Store tests, vet, and manifest checks pass.

Outcome:
- Corrected the `store/idempotency` package example to pass `dataidempotency.DefaultSQLConfig()` to `NewSQLStore`.
- Added a store behavior matrix covering missing, expired, closed, invalid key/path, nil context, and delete-missing semantics across store subpackages.
- Synced `store/file` README and testing guidance with stable path/error expectations.

Validation:
- go test -timeout 20s ./store/...
- go vet ./store/...
- go run ./internal/checks/module-manifests
