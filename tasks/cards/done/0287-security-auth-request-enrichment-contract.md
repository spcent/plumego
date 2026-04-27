# Card 0287: Security Auth Request Enrichment Contract

Priority: P1
State: done
Primary Module: security

## Goal

Make request-enriching authenticators an explicit `security/authn` contract instead of an implicit private interface owned by `middleware/auth`.

## Problem

Authentication contract ownership is split in a non-obvious way:

- `security/authn.Authenticator` owns the public authentication primitive.
- `middleware/auth` defines a private `requestAuthenticator` interface with `AuthenticateRequest(*http.Request) (*authn.Principal, *http.Request, error)`.
- `security/jwt.Authenticator` implements `AuthenticateRequest` implicitly so middleware can preserve enriched request context with token claims.
- `JWTManager.Authenticator(...)` returns `authn.Authenticator`, which hides the fact that the returned value may also enrich the request.

The capability is valid, but the ownership is backwards: `middleware/auth` should adapt a security contract, not define a hidden extension contract that security implementations accidentally satisfy.

## Scope

- Before editing, grep every `AuthenticateRequest`, `requestAuthenticator`, and `JWTManager.Authenticator` call site.
- Add an explicit request-enrichment authenticator contract under `security/authn`.
- Update `middleware/auth` to type-assert against the `security/authn` contract, not a middleware-local private duplicate.
- Update `security/jwt.Authenticator` and `JWTManager.Authenticator(...)` so the exported return type communicates the available capability clearly.
- Update tests that currently rely on the private middleware shape.
- Keep principal storage through `authn.WithPrincipal` / `authn.PrincipalFromContext`.

## Non-Goals

- Do not move HTTP middleware behavior into `security`.
- Do not add tenant policy, quota, or session lifecycle behavior to stable auth packages.
- Do not change JWT token verification algorithms or claim validation semantics.
- Do not introduce context service-locator behavior beyond explicit principal and token-claim accessors.

## Expected Files

- `security/authn/authn.go`
- `security/jwt/auth_jwt.go`
- `middleware/auth/contract.go`
- `middleware/auth/contract_test.go`
- `security/jwt/*_test.go`

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./security/... ./middleware/auth
go test -race -timeout 60s ./security/... ./middleware/auth
go vet ./security/... ./middleware/auth
```

Then run the required repo-wide gates before committing.

## Done Definition

- Request-enriching authenticators are declared in `security/authn`.
- `middleware/auth` contains only HTTP adapter and error-response behavior.
- `security/jwt` exposes its request-enrichment capability through the canonical auth contract.
- No private cross-package shadow contract remains in `middleware/auth`.
- Focused gates and repo-wide gates pass.

## Outcome

- Added `authn.RequestAuthenticator` to make request enrichment an explicit security contract.
- `middleware/auth` now type-asserts against `authn.RequestAuthenticator` instead of a private interface.
- `JWTManager.Authenticator` returns `authn.RequestAuthenticator` and the JWT authenticator doc comment reflects the contract.
