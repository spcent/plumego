# Card 2045

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Context Package: implementation
Priority: P0
State: active
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/app/config.go
- reference/workerfleet/internal/handler
- reference/workerfleet/README.md
- reference/workerfleet/docs/api.md
- reference/workerfleet/env.example
Depends On:

Goal:
Require explicit admin/API Bearer auth for workerfleet query endpoints in production.

Scope:
Add query auth config, wire it into query handlers, keep worker ingress auth separate, and update local env/docs/deployment guidance.

Non-goals:
- Do not add multi-user RBAC, sessions, JWT, or OIDC.
- Do not require auth for `/healthz` or `/readyz`.
- Do not move auth logic into Plumego stable roots.

Files:
- reference/workerfleet/internal/app/config.go
- reference/workerfleet/internal/handler/worker_register.go
- reference/workerfleet/internal/handler/routes.go
- reference/workerfleet/internal/handler/*_test.go
- reference/workerfleet/README.md

Acceptance Tests:
- reference/workerfleet/internal/app/config_test.go: TestLoadConfigRejectsProductionWithoutAdminAuthToken
- reference/workerfleet/internal/handler/routes_test.go: TestQueryRoutesRequireAdminAuthWhenConfigured

Tests:
- Existing handler and app config tests.
- Negative tests for missing, malformed, and invalid query auth credentials.

Docs Sync:
- reference/workerfleet/README.md
- reference/workerfleet/docs/api.md
- reference/workerfleet/env.example
- env.example

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/app ./internal/handler
- cd reference/workerfleet && go vet ./internal/app ./internal/handler
- gofmt -l reference/workerfleet/internal/app reference/workerfleet/internal/handler

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
