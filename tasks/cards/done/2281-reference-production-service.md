# Card 2281

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P1
State: active
Primary Module: reference
Owned Files:
- reference/production-service/README.md
- reference/production-service/main.go
- reference/production-service/internal/app/app.go
- reference/production-service/internal/app/routes.go
- reference/production-service/internal/config/config.go
Depends On: 2270, 2271

Goal:
Add a runnable production-oriented reference service that extends `reference/standard-service` with explicit security and observability wiring.

Scope:
- Create `reference/production-service` as an application example, not a reusable framework layer.
- Demonstrate explicit middleware ordering for request IDs, recovery, body limits, timeout, security headers, abuse guard, and request observability.
- Keep route registration visible in `internal/app/routes.go`.
- Keep config ownership app-local and standard-library-first.

Non-goals:
- Do not introduce a hidden production bundle.
- Do not mount `x/devtools` by default.
- Do not add new stable-root APIs or external dependencies.

Files:
- `reference/production-service/README.md`
- `reference/production-service/main.go`
- `reference/production-service/internal/app/app.go`
- `reference/production-service/internal/app/routes.go`
- `reference/production-service/internal/config/config.go`

Tests:
- `go test -timeout 20s ./reference/production-service/...`
- `go run ./internal/checks/reference-layout`
- `scripts/check-spec tasks/cards/done/2281-reference-production-service.md`

Docs Sync:
- Required because a new reference application is added.

Done Definition:
- `reference/production-service` builds and runs as a concrete production-oriented example.
- Security and observability decisions are visible in app-local wiring.
- `reference/standard-service` remains the minimal canonical learning path.

Outcome:
- Added `reference/production-service` as a runnable production-oriented
  reference application.
- Kept configuration app-local, route registration visible in
  `internal/app/routes.go`, and middleware ordering explicit in
  `internal/app/app.go`.
- Demonstrated request IDs, recovery, body limits, timeout, security headers,
  abuse guard, tracing hook, HTTP metrics, and access logs without adding a
  hidden production bundle.
- Kept `x/devtools` unmounted by default and left `reference/standard-service`
  as the minimal canonical learning path.

Validations:
- `go test -timeout 20s ./reference/production-service/...`
- `go run ./internal/checks/reference-layout`
