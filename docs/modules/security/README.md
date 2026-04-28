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

## Production Profile Relationship

Use `security/*` for reviewable primitives and policies:

- `security/authn` owns principals, authenticators, authorizers, and context accessors.
- `security/headers` owns header policies consumed by `middleware/security`.
- `security/headers` treats proxy TLS headers as HTTPS only when the whole relevant forwarded chain is explicitly secure.
- `security/input` owns input-safety checks and rejects unsafe HTTP header names or values before they reach transport adapters.
- `security/abuse` owns abuse guard decisions consumed by `middleware/ratelimit`.
- `security/abuse` reports limiter bucket metrics from the same accounting path used for eviction and cleanup decisions.
- `security/jwt` and `security/password` own token and password primitives.
- `security/jwt` verification fails closed when configured issuer, configured audience, or subject claims are missing or mismatched.
- `security/password` exposes sentinel errors for invalid cost, invalid stored hash, and password mismatch so callers can classify failures with `errors.Is`.

HTTP request wiring belongs in `middleware/auth`, `middleware/security`, and
`middleware/ratelimit`. Application-specific authorization decisions should be
constructor-injected into those adapters rather than hidden behind package
globals.
