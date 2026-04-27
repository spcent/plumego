# Card 0206

Priority: P1
State: done
Primary Module: security
Owned Files:
- `security/jwt/jwt.go`
- `security/headers/headers.go`
- `middleware/auth/contract.go`
- `middleware/auth/auth.go`
- `middleware/security/headers.go`
Depends On:

Goal:
- Remove overlapping stable auth/header middleware entrypoints and restore one explicit transport-adapter path on top of security primitives.
- Converge directly to one adapter surface per concern, without preserving twin APIs for compatibility.

Problem:
- Stable auth flow is currently split between `middleware/auth` and `security/jwt`: both expose middleware entrypoints, but they attach identity/context in different ways and invite multiple competing integration styles.
- JWT claims are stored with a package-local `context.WithValue` key and exposed only through `GetClaimsFromContext`, which does not follow the repo’s `With{Type}` + `{Type}FromContext` accessor convention.
- Security headers also have two public middleware entrypoints: `security/headers.Policy.Middleware()` and `middleware/security.SecurityHeaders(...)`, which duplicate the same transport adapter responsibility.

Scope:
- Pick one canonical stable adapter path for authentication/authorization middleware and one for security-header application.
- Add explicit, convention-compliant claims accessors where stable request-context storage remains necessary.
- Keep `security/*` focused on primitives and policies while `middleware/*` owns the transport-only adapter surface.
- Remove the losing public entrypoints in the same card rather than leaving deprecated wrappers behind.

Non-goals:
- Do not redesign JWT token contents, key rotation, or authorization policy semantics.
- Do not add tenant-aware auth behavior.
- Do not broaden middleware into a security policy catalog.
- Do not preserve duplicate public middleware entrypoints for migration convenience.

Files:
- `security/jwt/jwt.go`
- `security/headers/headers.go`
- `middleware/auth/contract.go`
- `middleware/auth/auth.go`
- `middleware/security/headers.go`

Tests:
- `go test -timeout 20s ./security/... ./middleware/auth ./middleware/security`
- `go test -race -timeout 60s ./security/... ./middleware/auth ./middleware/security`
- `go vet ./security/... ./middleware/auth ./middleware/security`

Docs Sync:
- Sync the security and middleware primers if the canonical adapter entrypoint changes or if claims/context access rules become more explicit.

Done Definition:
- Stable auth and security-header wiring each have one obvious public adapter path.
- JWT claims storage follows the repo accessor convention instead of ad hoc raw context keys.
- Duplicate public wrappers are removed in the same change that migrates their last callers.
- No second public adapter path remains for auth or security headers.

Outcome:
- Completed.
- Removed duplicate auth and security-header transport wrappers from stable `security`, keeping the canonical adapters in stable `middleware`.
- Converged JWT claims context access on explicit `WithTokenClaims` / `TokenClaimsFromContext` accessors.
