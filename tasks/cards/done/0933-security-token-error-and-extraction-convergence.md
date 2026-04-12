# Card 0933: Security Token Error + Extraction Convergence

Priority: P1
State: done
Primary Module: security

## Goal

Provide one canonical set of token-auth error signals and one canonical bearer-token extraction path across `security/authn` and `security/jwt`.

## Problem

`security/authn` defines `ErrInvalidToken` and `ErrExpiredToken`, while `security/jwt` defines its own `ErrInvalidToken` and `ErrTokenExpired` (same meaning, different values and names). `security/jwt` also has its own bearer-token extraction logic separate from `security/authn` static-token extraction. This duplicates behavior and makes error handling inconsistent across stable auth primitives.

## Scope

- Enumerate all exported error symbols in `security/authn` and `security/jwt` before edits.
- Decide canonical error values (prefer `security/authn` errors for transport-facing auth failures).
- Remove or alias duplicate JWT error symbols and update call sites/tests accordingly.
- Consolidate bearer-token extraction into a single helper (shared by static-token authn and JWT authn).
- Keep the “no query-string tokens” rule intact.
- Update any tests that assert on old error values or old extraction behavior.

## Non-Goals

- Do not change JWT signing, rotation, or claims behavior.
- Do not introduce new auth policy or multi-scheme negotiation.
- Do not move JWT or authn code out of `security`.

## Expected Files

- `security/authn/*.go`
- `security/jwt/*.go`
- `security/*/*_test.go`
- `docs/modules/security/README.md` (if error codes are documented)

## Validation

```bash
go test -timeout 20s ./security/...
go test -race -timeout 60s ./security/...
go vet ./security/...
```

## Done Definition

- Exactly one canonical invalid/expired token error is used across authn and JWT.
- JWT authn and static token authn share the same bearer-token extraction helper.
- All relevant tests pass.
- Any removed exported symbols have zero residual references.

## Outcome

- JWT invalid/expired token errors now alias the authn errors, unifying transport-facing error values.
- Added `authn.ExtractBearerToken` and routed JWT/static token auth to the shared helper.
- Updated JWT negative-matrix and extraction tests to expect canonical auth error codes.
- Validation: `go test -timeout 20s ./security/...`, `go test -race -timeout 60s ./security/...`, `go vet ./security/...`.
