# Card 0025

Priority: P1

Goal:
- Add focused route-level tests that lock down `RegisterContextResourceRoutes(...)` and related public route registration behavior.

Scope:
- `x/rest` route registration tests
- public route surface only

Non-goals:
- Do not redesign controller internals in this card.
- Do not change resource query behavior in this card.

Files:
- `x/rest/entrypoints.go`
- `x/rest/routes_test.go`
- `x/rest/README.md`

Tests:
- `go test ./x/rest/...`

Docs Sync:
- Update `x/rest/README.md` only if the tested route surface needs clearer explanation.

Done Definition:
- Public route registration behavior is covered by focused tests.
- Future route-surface regressions fail in `x/rest` tests before spreading further.
