# Card 1421

Milestone: v1-breaking-normalization
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: active
Primary Module: contract
Owned Files:
- contract/*
- docs/modules/contract/README.md
- README.md
- README_CN.md
Depends On:
- none

Goal:
- Remove compatibility response, error, and binding surfaces from `contract` so
  v1 has one canonical transport contract.

Scope:
- Enumerate exported compatibility symbols and response/error envelope fields.
- Remove deprecated or duplicate helpers after migrating all callers.
- Keep `WriteResponse` as the only success response path.
- Keep `WriteError` plus `NewErrorBuilder` as the only error response path.
- Normalize request metadata/context accessors to one `With{Type}` and
  `{Type}FromContext` shape.
- Update focused tests and migration documentation for breaking removals.

Non-goals:
- Do not add tracing, tenant, session, observability, or business concerns.
- Do not introduce non-stdlib dependencies.
- Do not change handler signatures away from `net/http`.

Files:
- contract/response.go
- contract/errors.go
- contract/error_codes.go
- contract/context_core.go
- contract/*_test.go

Tests:
- go test -timeout 20s ./contract
- go vet ./contract
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update affected README/docs if public symbols or envelope semantics change.
- Add migration notes for removed compatibility symbols.

Done Definition:
- Removed symbols have no remaining Go callers.
- Contract tests cover the canonical response and error paths.
- Dependency rules pass.

Outcome:

