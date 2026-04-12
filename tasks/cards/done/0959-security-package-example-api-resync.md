# Card 0959: Security Package Example API Resync

Priority: P2
State: done
Primary Module: security

## Goal

Align stable `security/*` package docs and examples with the real public API so
readers see the current canonical helpers instead of stale pre-rename or
pre-refactor symbols.

## Problem

- `security/password/password.go` package comments still show old example calls:
  - `password.Hash(...)`
  - `password.Verify(...)`
  - `password.ValidateStrength(...)`
- `security/headers/headers.go` package-level example still uses stale policy
  field names and shapes such as:
  - `XFrameOptions`
  - `ContentSecurityPolicy: &headers.CSPOptions{...}`
- `security/jwt/jwt.go` package-level examples still teach a pre-refactor API:
  - `jwt.NewManager(...)`
  - `jwt.Claims`
  - `manager.Generate(...)`
  - `manager.Verify(...)`
  - `kv.New()`
- The real exported stable API is already different:
  - password uses `HashPassword(...)`, `CheckPassword(...)`,
    `ValidatePasswordStrength(...)`
  - headers uses `Policy{FrameOptions, ContentTypeOptions, ContentSecurityPolicy string, ...}`
    plus `NewCSPBuilder()` / `StrictCSP()`
  - jwt uses `NewJWTManager(...)`, `GenerateTokenPair(...)`,
    `VerifyToken(...)`, and `store/kv.NewKVStore(...)`

- Tests already target the current API, so the drift is documentation-only but
  directly affects the stable learning surface.

This is exactly the kind of stale wrapper-era naming drift that makes the
stable security surface feel less canonical than it actually is.

## Scope

- Rewrite package-level examples/comments in stable `security/*` packages to use
  the live exported names and field shapes.
- Cover at least `security/password`, `security/headers`, and `security/jwt`.
- Check nearby stable security docs for the same stale symbols and fix them if
  present.
- Keep examples fail-closed and aligned with the current return types and
  constructor paths.

## Non-Goals

- Do not reintroduce deprecated wrapper aliases like `Hash`, `Verify`,
  `ValidateStrength`, `NewManager`, `Generate`, or `Verify`.
- Do not change password hashing behavior, JWT behavior, or header policy
  semantics in this card.
- Do not widen the card into a broader auth/security refactor.

## Files

- `security/password/password.go`
- `security/headers/headers.go`
- `security/jwt/jwt.go`
- `docs/modules/security/README.md`
- any stable docs/examples that still mention the stale helper names or stale
  field shapes

## Tests

- `go test -timeout 20s ./security/password ./security/headers ./security/jwt`
- `go test -race -timeout 60s ./security/password ./security/headers ./security/jwt`
- `go vet ./security/password ./security/headers ./security/jwt`
- `rg -n 'password\\.(Hash|Verify|ValidateStrength)\\(|jwt\\.(NewManager|Claims)\\b|manager\\.(Generate|Verify)\\(|kv\\.New\\(|XFrameOptions|CSPOptions' security docs/modules README.md README_CN.md`

## Docs Sync

- No repo-wide doc sync expected beyond touched stable security docs unless the
  stale examples appear elsewhere.

## Done Definition

- Stable security package docs no longer reference removed or stale helper
  names, constructors, or field shapes.
- Examples use the current exported API and match real function signatures.
- Grep for the stale example symbols is empty in the touched stable docs.

## Outcome

- Updated `security/password` package examples to use
  `HashPassword(...)`, `CheckPassword(...)`, and
  `ValidatePasswordStrength(...)`, and corrected the feature summary to match
  the implementation.
- Updated the `security/headers` package-level example to use the real
  `Policy` fields plus `NewCSPBuilder()`.
- Updated `security/jwt` package examples to use `NewJWTManager(...)`,
  `GenerateTokenPair(...)`, `VerifyToken(...)`, and `store/kv.NewKVStore(...)`.

## Validation Run

```bash
go test -timeout 20s ./security/password ./security/headers ./security/jwt
go test -race -timeout 60s ./security/password ./security/headers ./security/jwt
go vet ./security/password ./security/headers ./security/jwt
rg -n 'password\.(Hash|Verify|ValidateStrength)\(|jwt\.(NewManager|Claims)\b|manager\.(Generate|Verify)\(|kv\.New\(|XFrameOptions|CSPOptions' security docs/modules README.md README_CN.md
```
