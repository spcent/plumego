# Card 0715

Milestone: —
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: x/gateway
Owned Files:
- `x/gateway/proxy.go`
- `x/gateway/proxy_test.go`
Depends On: —

Goal:
- Prevent the gateway proxy from retrying or writing an error response after it has already committed a downstream response to the client.

Problem:
`Proxy.proxyRequest` writes upstream headers and status before copying the body. If `copyResponse` fails after `WriteHeader`, `handleHTTP` treats the error like a retryable proxy failure and may retry or call `ErrorHandler`, even though the client response has already started.

Scope:
- Track whether a downstream response has been committed for a proxy attempt.
- Treat body-copy failures after commit as terminal for that client response.
- Avoid calling `ErrorHandler` after headers/body are already committed.
- Add regression coverage for copy failure before commit versus after commit.

Non-goals:
- Do not redesign the gateway load balancer.
- Do not change backend health/circuit-breaker policy except where commit state requires it.
- Do not add non-stdlib proxy dependencies.

Files:
- `x/gateway/proxy.go`
- `x/gateway/proxy_test.go`

Tests:
- `go test -timeout 20s ./x/gateway/...`
- `go vet ./x/gateway/...`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Not required unless public retry semantics are documented elsewhere during the fix.

Done Definition:
- Once response headers are committed, gateway does not retry the request or invoke the error writer for that client response.
- Pre-commit proxy failures still use the existing retry/error path.
- Tests lock both paths.

Outcome:
- Added proxy attempt commit-state tracking for HTTP proxy responses.
- Body copy errors after `WriteHeader` now terminate the request without retrying or invoking `ErrorHandler`.
- Pre-commit failures still use the existing retry and final error-handler path.
- Added regression tests for both committed copy failures and pre-commit response modification failures.
- Validation passed:
  - `go test -timeout 20s ./x/gateway/...`
  - `go vet ./x/gateway/...`
  - `go run ./internal/checks/dependency-rules`
