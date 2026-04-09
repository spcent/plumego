# security

## Purpose

`security` contains reviewable authentication, header, input-safety, and abuse-guard primitives.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- verifying tokens or signatures
- enforcing security headers
- validating hostile or malformed input

## Do not use this module for

- app bootstrap
- business-specific authorization policy hidden in middleware
- logging secrets or tokens

## First files to read

- `security/module.yaml`
- the target package under `security/*`
- `AGENTS.md` security rules

## Canonical change shape

- fail closed
- keep verification explicit
- add negative tests for invalid input and invalid credentials
- keep `Principal`, `Authenticator`, `Authorizer`, and the canonical `WithPrincipal(...)` / `PrincipalFromContext(...)` accessors in `security/authn`
- keep JWT, header, and signature logic in `security/*` as primitives and policies
- keep session revocation, token-version invalidation, and tenant-session sentinel errors in `x/tenant/session`, not in stable `security/*`
- keep reusable resilience primitives such as circuit breakers in `x/resilience`, not in stable `security/*`
- route HTTP adapter wiring through `middleware/auth` and `middleware/security`
