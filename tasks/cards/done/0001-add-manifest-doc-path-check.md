# Card 0001

Priority: P0

Goal:
- Extend manifest validation so every declared `doc_paths` target must exist.

Scope:
- `internal/checks/module-manifests`
- shared check helpers if needed

Non-goals:
- Do not add new module docs in this card.
- Do not expand validation beyond declared `doc_paths`.

Files:
- `internal/checks/module-manifests/main.go`
- `internal/checks/checkutil/checkutil.go`
- `internal/checks/checkutil/checkutil_test.go`

Tests:
- `go run ./internal/checks/module-manifests`
- `go test ./internal/checks/checkutil/...`

Docs Sync:
- Update `AGENTS.md` only if the checker name or required gate list changes.

Done Definition:
- A manifest with a missing `doc_paths` target fails validation.
- Existing valid manifests continue to pass.
- The new rule is enforced by current manifest validation flow rather than tribal knowledge.
