# Card 2282

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P2
State: active
Primary Module: reference
Owned Files:
- reference/with-rest/README.md
- reference/with-rest/main.go
- reference/with-tenant/README.md
- reference/with-tenant/main.go
- docs/README.md
Depends On: 2281

Goal:
Add runnable REST and tenant scenario references that start from the canonical app shape and add one capability family explicitly.

Scope:
- Add `reference/with-rest` showing `x/rest` resource wiring without replacing normal handler routes.
- Add `reference/with-tenant` showing tenant resolution plus policy/quota/rate-limit composition in a small API.
- Keep both examples offline and standard-library runnable.
- Link the scenarios from the docs entrypoint.

Non-goals:
- Do not promote `x/rest` or `x/tenant` to beta.
- Do not create a second canonical app layout.
- Do not add database or network service dependencies.

Files:
- `reference/with-rest/README.md`
- `reference/with-rest/main.go`
- `reference/with-tenant/README.md`
- `reference/with-tenant/main.go`
- `docs/README.md`

Tests:
- `go test -timeout 20s ./reference/with-rest ./reference/with-tenant`
- `go run ./internal/checks/reference-layout`
- `scripts/check-spec tasks/cards/done/2282-reference-rest-tenant-scenarios.md`

Docs Sync:
- Required because new scenario references are added.

Done Definition:
- Users can run REST and tenant scenarios without external services.
- Both examples preserve explicit app wiring and canonical handler conventions.
- Scenario docs state that the used `x/*` modules remain experimental unless separately promoted.

Outcome:
