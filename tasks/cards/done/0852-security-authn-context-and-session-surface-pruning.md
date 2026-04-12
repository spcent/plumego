# Card 0852

Priority: P1
State: active
Primary Module: security
Owned Files:
- `security/authn/authn.go`
- `security/authn/authn_test.go`
- `middleware/auth`
- `x/tenant/session`
- `x/tenant/resolve`
- `security/module.yaml`
- `docs/modules/security/README.md`
- `docs/modules/x-tenant/README.md`

Goal:
- Keep stable `security/authn` focused on identity primitives and canonical context accessors only.
- Move session-lifecycle sentinel errors to `x/tenant/session` and remove duplicate request convenience wrappers.

Problem:
- `security/authn` still exports `ErrSessionRevoked`, `ErrSessionExpired`, `ErrRefreshReused`, and `ErrTokenVersionMismatch`, even though session lifecycle management belongs to `x/tenant/session`.
- `PrincipalFromRequest(...)` and `RequestWithPrincipal(...)` duplicate the canonical `WithPrincipal(...)` / `PrincipalFromContext(...)` accessor model and encourage extra helper surface around `*http.Request`.
- `x/tenant/session` already owns session lifecycle contracts, but stable `security/authn` still leaks tenant-session errors and request wrappers into middleware and extensions.

Scope:
- Remove session lifecycle error sentinels from stable `security/authn` and move them to `x/tenant/session`.
- Remove `PrincipalFromRequest(...)` and `RequestWithPrincipal(...)` so `security/authn` exposes only the canonical context accessor pair.
- Update `middleware/auth`, `x/tenant/session`, `x/tenant/resolve`, and tests in the same change.
- Sync docs/manifests to the reduced auth primitive boundary.

Non-goals:
- Do not redesign authenticator/authorizer interfaces.
- Do not redesign tenant session middleware behavior.
- Do not add compatibility aliases in either stable or x/* packages.

Files:
- `security/authn/authn.go`
- `security/authn/authn_test.go`
- `middleware/auth`
- `x/tenant/session`
- `x/tenant/resolve`
- `security/module.yaml`
- `docs/modules/security/README.md`
- `docs/modules/x-tenant/README.md`

Tests:
- `go test -timeout 20s ./security/authn ./middleware/auth ./x/tenant/session ./x/tenant/resolve`
- `go test -race -timeout 60s ./security/authn ./middleware/auth ./x/tenant/session ./x/tenant/resolve`
- `go vet ./security/authn ./middleware/auth ./x/tenant/session ./x/tenant/resolve`

Docs Sync:
- Keep security docs aligned on the rule that stable `authn` owns identity primitives only.
- Keep tenant docs aligned on the rule that session lifecycle errors live in `x/tenant/session`.

Done Definition:
- Stable `security/authn` exports only auth primitives plus the canonical `WithPrincipal(...)` / `PrincipalFromContext(...)` accessors.
- Session lifecycle errors live in `x/tenant/session`.
- There are zero residual references to removed authn request wrappers or session error sentinels.

Outcome:
- Completed.
- Removed `PrincipalFromRequest(...)` and `RequestWithPrincipal(...)` from stable `security/authn`, leaving `WithPrincipal(...)` and `PrincipalFromContext(...)` as the canonical accessors.
- Moved session lifecycle sentinel errors to `x/tenant/session` and updated tenant/session validators, middleware, and JWT-backed state checks to use the new owner-side errors.
- Reduced stable `middleware/auth` back to stable auth error handling only, and updated security/tenant docs to reflect the narrower ownership split.
