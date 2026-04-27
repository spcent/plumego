# Card 0026

Priority: P1

Goal:
- Add a canonical example showing `ResourceSpec -> repository -> RegisterContextResourceRoutes(...)`.

Scope:
- `x/rest` example or example-style test
- docs for the canonical reuse path

Non-goals:
- Do not build a full sample service under `reference/`.
- Do not introduce new runtime abstractions.

Files:
- `x/rest/example_test.go`
- `x/rest/README.md`
- `docs/ROADMAP.md`

Tests:
- `go test ./x/rest/...`

Docs Sync:
- Keep the example aligned with the README and roadmap wording for Phase 3.

Done Definition:
- `x/rest` has one obvious example for reusable resource API composition.
- New contributors can follow one canonical path without inventing local scaffolding.
