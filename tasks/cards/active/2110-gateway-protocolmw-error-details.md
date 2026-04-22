# Card 2110: Gateway Protocol Middleware Error Details

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: high
State: active
Primary Module: x/gateway
Owned Files:
- x/gateway/protocolmw/middleware.go
- x/gateway/protocolmw/middleware_test.go
- docs/modules/x-gateway/README.md
Depends On: none

Goal:
- Stop protocol middleware from exposing raw adapter, transform, executor, encoder, or request-read errors in structured response details.
- Keep gateway protocol failures explicit, fail-closed, and diagnosable through stable codes rather than internal `cause` strings.

Scope:
- Audit `BuildHTTPHandler`, `WrapHTTPHandler`, and protocol middleware error details.
- Replace `map[string]any{"cause": err.Error()}` style client details with safe detail values or omit details.
- Preserve custom `ErrorMapper`, `LogError`, and protocol adapter extension points.
- Add tests that assert unsafe internal error text is not present in HTTP response bodies.

Non-goals:
- Do not change reverse proxy, cache, balancing, or rewrite behavior.
- Do not add tenant, business policy, or bootstrap ownership to `x/gateway`.
- Do not remove caller-provided custom error mapping hooks.

Files:
- `x/gateway/protocolmw/middleware.go`: normalize protocol middleware error detail construction.
- `x/gateway/protocolmw/middleware_test.go`: cover transform, executor, encode, and read failure responses.
- `docs/modules/x-gateway/README.md`: document safe protocol middleware error details if behavior changes.

Tests:
- `go test -race -timeout 60s ./x/gateway/...`
- `go test -timeout 20s ./x/gateway/...`
- `go vet ./x/gateway/...`

Docs Sync:
- Required if response detail semantics or documented protocol middleware behavior change.

Done Definition:
- Protocol middleware no longer emits raw `err.Error()` values in default client-facing details.
- Custom error mapping remains supported and covered by tests.
- The three listed validation commands pass.

Outcome:

