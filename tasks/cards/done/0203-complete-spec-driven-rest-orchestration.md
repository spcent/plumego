# Card 0203

Priority: P1

Goal:
- Push `ResourceSpec` further toward a single orchestration surface for query, hooks, transformer, and controller defaults.

Scope:
- `x/rest/spec.go`
- `x/rest/resource.go`
- `x/rest/resource_db.go`

Non-goals:
- Do not redesign the repository abstraction.
- Do not add gateway or bootstrap concerns.

Files:
- `x/rest/spec.go`
- `x/rest/resource.go`
- `x/rest/resource_db.go`
- `x/rest/spec_test.go`

Tests:
- `go test ./x/rest/...`

Docs Sync:
- Update `x/rest/README.md` if the recommended spec-driven path changes.

Done Definition:
- `ResourceSpec` remains the dominant configuration surface.
- Controller defaults no longer drift across ad hoc initialization paths.
- Spec-driven behavior is fixed by tests.
