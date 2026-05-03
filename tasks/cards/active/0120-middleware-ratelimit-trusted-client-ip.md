# Card 0120

Milestone:
Recipe: specs/change-recipes/security.yaml
Priority: P0
State: active
Primary Module: middleware
Owned Files:
  - middleware/internal/transport/http.go
  - middleware/internal/transport/transport_test.go
  - middleware/ratelimit/abuse_guard.go
  - middleware/ratelimit/abuse_guard_test.go
  - docs/modules/middleware/README.md
Depends On: 0119

Goal:
Make rate-limit key extraction safe by default. The current default trusts
`X-Forwarded-For` and `X-Real-IP`, which is unsafe unless the application is
behind a trusted proxy.

Scope:
- Add a transport helper that uses `RemoteAddr` only.
- Change `ratelimit.AbuseGuard` default `KeyFunc` to the direct remote address helper.
- Document how callers can opt into proxy-aware extraction explicitly.
- Preserve the existing proxy-aware helper for logs and explicit caller use.

Non-goals:
- Do not implement a full trusted-proxy CIDR policy in this card.
- Do not change access-log client IP behavior.
- Do not change `security/abuse` limiter primitives.

Files:
- `middleware/internal/transport/http.go`
- `middleware/internal/transport/transport_test.go`
- `middleware/ratelimit/abuse_guard.go`
- `middleware/ratelimit/abuse_guard_test.go`
- `docs/modules/middleware/README.md`

Tests:
- `go test -timeout 20s ./middleware/internal/transport ./middleware/ratelimit`
- `go test -race -timeout 60s ./middleware/ratelimit`
- `go vet ./middleware/...`

Docs Sync:
- `docs/modules/middleware/README.md`

Done Definition:
- Default rate-limit key cannot be spoofed through forwarded headers.
- Docs describe explicit proxy-aware override.
- Targeted tests pass.

Outcome:

