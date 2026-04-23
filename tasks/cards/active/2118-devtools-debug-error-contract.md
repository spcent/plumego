# Card 2118: Devtools Debug Error Contract

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: high
State: active
Primary Module: x/devtools
Owned Files:
- x/devtools/devtools.go
- x/devtools/devtools_test.go
- x/devtools/pubsubdebug/component.go
- x/devtools/pubsubdebug/configure.go
- x/devtools/pubsubdebug/component_test.go
Depends On: none

Goal:
- Normalize devtools debug-route error responses so debug endpoints still follow the project HTTP error contract.
- Replace lowercase ad hoc codes and raw reload error messages with stable, safe responses.

Scope:
- Audit `x/devtools/devtools.go` env reload handler errors.
- Audit pubsub debug component/configuration handlers for lowercase `not_supported` codes and missing explicit codes.
- Keep devtools opt-in, debug-only, and subordinate to the observability family.
- Add focused tests for reload failure and unsupported pubsub debug behavior.

Non-goals:
- Do not move devtools routes into `core`, stable observability roots, or production admin policy.
- Do not change profiling, metrics, runtime snapshot, or route registration semantics.
- Do not add hidden runtime side effects or global debug registration.

Files:
- `x/devtools/devtools.go`: use explicit stable code and safe message for env reload failures.
- `x/devtools/devtools_test.go`: cover env reload error responses.
- `x/devtools/pubsubdebug/component.go`: normalize component-level unsupported and nil-dependency errors.
- `x/devtools/pubsubdebug/configure.go`: normalize configure-route unsupported errors.
- `x/devtools/pubsubdebug/component_test.go`: cover unsupported and nil pubsub debug errors.

Tests:
- `go test -race -timeout 60s ./x/devtools/...`
- `go test -timeout 20s ./x/devtools/...`
- `go vet ./x/devtools/...`

Docs Sync:
- Required only if `docs/modules/x-devtools/README.md` documents exact debug error codes or messages.

Done Definition:
- Devtools debug HTTP errors in scope use explicit stable codes.
- Lowercase `env_reload_failed` and `not_supported` response codes are removed from active handler paths.
- Reload failures do not expose raw underlying error strings as public response messages.
- The three listed validation commands pass.

Outcome:

