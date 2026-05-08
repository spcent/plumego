# 0903 - x/websocket token helper hardening

Status: done
Priority: P1
Primary module: `x/websocket`

## Problem

The built-in HS256 token helper does not validate secret strength and ignores
malformed temporal claims. Stable usage must make the helper's limited contract
explicit and fail closed on malformed inputs.

## Scope

- Validate HS256 secrets before constructing token auth.
- Reject malformed `exp` claims rather than ignoring them.
- Keep issuer, audience, `nbf`, and `iat` outside the built-in helper unless
  explicitly configured.
- Update tests and docs to call this a simple HS256 verifier, not full JWT/OIDC
  policy.

## Out of Scope

- Full OIDC/JWKS implementation.
- New non-standard-library dependencies.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/dependency-rules`

## Outcome

- Changed `NewHS256TokenAuth` to fail visibly with `ErrWeakJWTSecret` when the secret is shorter than 32 bytes.
- Reused the same secret-strength validation in websocket server config and secure room auth construction.
- Rejected malformed `exp` claims instead of silently ignoring string, fractional, or negative values.
- Updated tests and docs to present the helper as a lightweight HS256 verifier, not a full OIDC/JWT policy engine.

Completed validations:

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/dependency-rules`
- `go build ./...`

Not run:

- `bash scripts/check-doc-snippets-compile.sh` because that script is not present in this checkout.
