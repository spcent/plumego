# 6208 - internal/nethttp SSRF host canonicalization

State: done
Priority: P0
Primary module: `internal/nethttp`

## Goal

Make SSRF host policy checks apply to canonical hostnames and reject URLs that
use userinfo syntax to obscure the authority.

## Scope

- Normalize hostnames for allowlist and blocklist comparisons.
- Ensure trailing-dot host variants do not bypass configured host policy.
- Reject URLs containing userinfo before DNS resolution.

## Non-goals

- Do not add new dependencies.
- Do not expand outbound client behavior beyond validation.
- Do not change default scheme or private IP policy.

## Files

- `internal/nethttp/security.go`
- `internal/nethttp/security_test.go`

## Tests

- `go test ./internal/nethttp`
- `go test ./internal/...`

## Docs Sync

No docs sync expected; this tightens internal validation behavior without adding
configuration.

## Done Definition

- Blocklist and allowlist checks treat trailing-dot hostnames consistently.
- URLs with userinfo are rejected as SSRF risk.
- Existing SSRF validation tests continue to pass.

## Outcome

- Added canonical host comparison for SSRF allowlist and blocklist checks.
- Rejected userinfo URLs before DNS resolution.
- Validation:
  - `go test ./internal/nethttp`
  - `go test ./internal/...`
