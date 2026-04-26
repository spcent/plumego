# Card 2293

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P2
State: done
Primary Module: cmd/plumego
Owned Files:
- cmd/plumego/internal/scaffold/scaffold.go
- cmd/plumego/internal/scaffold/scaffold_test.go
- cmd/plumego/README.md
- docs/getting-started.md
- tasks/cards/active/README.md
Depends On: 2284, 2292

Goal:
Turn the `rest-api` and `tenant-api` scaffold profiles from profile markers
into runnable vertical templates.

Scope:
- Generate runnable route files for REST and tenant profiles.
- Keep generated code standard-library and Plumego-only.
- Add scaffold tests that compile generated output expectations.

Non-goals:
- Do not add database dependencies.
- Do not generate experimental subpackage imports outside the selected profile.
- Do not change the default scaffold profile.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/README.md`
- `docs/getting-started.md`
- `tasks/cards/active/README.md`

Tests:
- `cd cmd/plumego && go test -timeout 20s ./internal/scaffold/...`
- `cd cmd/plumego && go test -timeout 20s ./commands/...`
- `scripts/check-spec tasks/cards/done/2293-scenario-scaffold-runnable-rest-tenant.md`

Docs Sync:
- Required because scaffold output changes.

Done Definition:
- `rest-api` and `tenant-api` generated projects have runnable scenario routes.
- Default scaffold remains stable-root-only.

Outcome:
- Kept `rest-api` on the canonical scaffold with runnable `x/rest` users
  resource routes.
- Upgraded `tenant-api` generated routes to include a runnable
  resolve -> policy -> quota -> ratelimit chain and `/api/models` handler.
- Updated scaffold tests and docs to distinguish runnable scenario routes from
  remaining capability marker profiles.

Validations:
- `cd cmd/plumego && go test -timeout 20s ./internal/scaffold/...`
- `cd cmd/plumego && go test -timeout 20s ./commands/...`
- `scripts/check-spec tasks/cards/done/2293-scenario-scaffold-runnable-rest-tenant.md`
