# Card 2132: NetHTTP SSRF CIDR Initialization Contract
Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: active
Primary Module: internal/nethttp
Owned Files:
- internal/nethttp/security.go
- internal/nethttp/security_test.go
Depends On: none

Goal:
Remove the `init()` + package-global mutable CIDR population path from
`internal/nethttp` SSRF protection. The current private IP list is populated at
package initialization and panics if a built-in CIDR cannot parse, which is more
implicit than the rest of the security helper surface and violates the repo
preference against hidden initialization behavior.

Scope:
- Replace `init()` with deterministic, explicitly testable CIDR construction.
- Avoid runtime panics for built-in CIDR parsing on normal package import.
- Preserve existing SSRF behavior for private, reserved, loopback, link-local,
  multicast, and public IP cases.
- Keep the implementation dependency-free.

Non-goals:
- Do not change the public `SSRFProtection` configuration shape.
- Do not weaken private IP or DNS-resolution checks.
- Do not move SSRF protection into stable `security`; this card is only about
  the existing internal net/http helper.

Files:
- `internal/nethttp/security.go`
- `internal/nethttp/security_test.go`

Tests:
- go test -race -timeout 60s ./internal/nethttp
- go test -timeout 20s ./internal/nethttp
- go vet ./internal/nethttp

Docs Sync:
No docs update required unless the implementation changes exported behavior.

Done Definition:
- `internal/nethttp/security.go` has no `init()` function.
- Private/reserved/public IP test coverage still passes.
- Built-in CIDR construction is covered by a focused test.
- The listed validation commands pass.

Outcome:
