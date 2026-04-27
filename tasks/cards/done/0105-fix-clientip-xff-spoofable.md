# Card 0105

Priority: P2

Goal:
- Fix `clientIPFromRequest` so it does not unconditionally trust the
  `X-Forwarded-For` first value, which can be injected by a malicious client.

Problem:

`context_core.go:396-400`:
```go
func clientIPFromRequest(r *http.Request) string {
    if ip := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); ip != "" {
        return ip   // ← first value, set by the client
    }
    // ...
}
```

`X-Forwarded-For` is appended by each proxy as a request travels through the
network. The header format is:
```
X-Forwarded-For: <client>, <proxy1>, <proxy2>
```
The **first** value is whatever the original sender claims their IP is — this
is completely under client control and trivially spoofed:
```
curl -H "X-Forwarded-For: 1.2.3.4" https://api.example.com/
```

Code that uses `Ctx.ClientIP` for rate limiting, access control, audit logging,
or geographic routing will operate on the spoofed value.

The **last** (rightmost) value that your own infrastructure added is trustworthy
when you control the load balancer. In practice, the right approach depends on
deployment topology:

- **Single trusted proxy**: use the last value from `X-Forwarded-For`.
- **N trusted proxies**: use the Nth value from the right.
- **No trusted proxies**: use `RemoteAddr` only; ignore `X-Forwarded-For`.

Fix (two-part):

**Part 1**: Change `clientIPFromRequest` to use the **last** non-empty value
from `X-Forwarded-For` as the default. This is safer for the common
single-proxy deployment:
```go
parts := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
for i := len(parts) - 1; i >= 0; i-- {
    if ip := strings.TrimSpace(parts[i]); ip != "" {
        return ip
    }
}
```

**Part 2**: Add a `TrustedProxies int` field to `RequestConfig` (default 1)
so callers can configure how many rightmost proxies to skip. Document the
security implications in the `RequestConfig` godoc.

Note: Part 2 is optional for this card if it expands scope; the primary fix
is Part 1 (stop using first value).

Non-goals:
- Do not add CIDR-range proxy trust (that is a larger feature).
- Do not change the `X-Real-IP` fallback logic.

Files:
- `contract/context_core.go`
- (If Part 2) `contract/context_core.go` (RequestConfig struct)

Tests:
- Add a test: a request with `X-Forwarded-For: spoofed, real` should return
  `real`, not `spoofed`.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `clientIPFromRequest` no longer returns the first (client-supplied) XFF value.
- The security behaviour is documented in a comment on the function.
- Tests verify spoofed first values are not used.
- All tests pass.

Outcome:
- Completed in the 2026-04-05 contract cleanup batch.
- Verified as part of the shared contract/task-card completion pass.

Validation Run:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
- `go build ./...`
- `go test -timeout 20s ./...`
- `go test -race -timeout 60s ./...`
- `go vet ./...`
