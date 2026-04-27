# Card 0642

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: internal/httpx
Owned Files: internal/httpx/client_ip.go, internal/httpx/client_ip_test.go
Depends On:

Goal:
Avoid returning malformed client IP header values from `ClientIP`.

Scope:
- Validate `X-Forwarded-For` entries before returning them.
- Validate `X-Real-IP` before returning it.
- Fall back to `RemoteAddr` when forwarded headers are empty or malformed.
- Add focused malformed-header and IPv6 coverage.

Non-goals:
- Do not add trusted proxy configuration.
- Do not change the public function signature.
- Do not parse CIDR or proxy policy in this helper.

Files:
- internal/httpx/client_ip.go
- internal/httpx/client_ip_test.go

Tests:
- go test ./internal/httpx
- go test ./internal/...

Docs Sync:
- None; behavior becomes safer without public docs changes.

Done Definition:
- Malformed forwarded headers are ignored.
- Valid IPv4 and IPv6 values are preserved.
- Focused and internal package tests pass.

Outcome:
- Validated forwarded client IP header values before returning them.
- Added malformed header fallback and IPv6 coverage.
- Validation: `go test ./internal/httpx`; `go test ./internal/...`.
