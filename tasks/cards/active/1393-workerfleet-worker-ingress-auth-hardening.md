# Card 1393

Milestone: workerfleet-hardening
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: active
Primary Module: reference/workerfleet/internal/handler
Owned Files:
- reference/workerfleet/internal/app/config.go
- reference/workerfleet/internal/app/config_test.go
- reference/workerfleet/internal/handler/worker_register.go
- reference/workerfleet/internal/handler/worker_heartbeat.go
- reference/workerfleet/internal/handler/query_test.go
Depends On:
- 1392

Goal:
- Fail closed on worker registration and heartbeat ingress when production auth is configured.

Scope:
- Add app-local worker ingress auth configuration using standard-library primitives only.
- Prefer a minimal shared-secret or HMAC request signature design with timing-safe comparison.
- Apply the check to `POST /v1/workers/register` and `POST /v1/workers/heartbeat`.
- Keep local development behavior explicit: auth disabled only when no secret/signing key is configured.
- Add negative tests for missing, malformed, and invalid credentials.
- Ensure secrets are never logged or echoed in responses.

Non-goals:
- Do not add OAuth/JWT dependencies.
- Do not build mTLS termination in the app process unless explicitly selected before execution.
- Do not change query endpoints in this card.

Files:
- reference/workerfleet/internal/app/config.go
- reference/workerfleet/internal/app/config_test.go
- reference/workerfleet/internal/handler/worker_register.go
- reference/workerfleet/internal/handler/worker_heartbeat.go
- reference/workerfleet/internal/handler/query_test.go

Tests:
- cd reference/workerfleet && go test -timeout 20s ./internal/app/...
- cd reference/workerfleet && go test -timeout 20s ./internal/handler/...
- cd reference/workerfleet && go test -timeout 20s ./...

Docs Sync:
- Update `reference/workerfleet/README.md`, `reference/workerfleet/docs/api.md`, and `env.example` if new environment variables are added.

Done Definition:
- Register and heartbeat fail closed when auth is configured and credentials are absent or invalid.
- Secret comparison uses timing-safe comparison.
- No secret value appears in logs, errors, or JSON responses.
- Negative auth tests pass.

Outcome:
