# Card 0848

Priority: P1
State: done
Primary Module: security
Owned Files:
- `security/jwt/jwt.go`
- `security/module.yaml`
- `docs/modules/security/README.md`
- `x/tenant/session`

Goal:
- Remove session-lifecycle and token-revocation ownership from stable `security/jwt`.
- Keep stable JWT focused on token signing, parsing, and verification primitives while moving revocation/version state to the owning tenant/session layer.

Problem:
- `security/module.yaml` says session lifecycle management and revocation do not belong in stable `security`, but `security/jwt.JWTManager` still owns blacklist state, identity version invalidation, and rotation-backed key persistence.
- `RevokeToken(...)`, `IncrementIdentityVersion(...)`, blacklist/version prefixes, and the backing-store semantics are session lifecycle behavior rather than JWT verification primitives.
- `x/tenant/session` already owns session lifecycle contracts, including revocation and token-version semantics, but stable `security/jwt` still carries a parallel stateful lifecycle path.

Scope:
- Remove revocation/version lifecycle ownership from stable `security/jwt`.
- Move the stateful token/session invalidation behavior to `x/tenant/session` or another owning tenant-session package.
- Keep stable `security/jwt` focused on claims, signing keys, token generation, and verification.
- Update tests and docs in the same change so no callers rely on the removed stable lifecycle API.

Non-goals:
- Do not redesign auth middleware adapters.
- Do not add compatibility wrappers in `security/jwt`.
- Do not move basic token parsing/signature verification out of stable `security`.

Files:
- `security/jwt/jwt.go`
- `security/module.yaml`
- `docs/modules/security/README.md`
- `x/tenant/session`

Tests:
- `go test -timeout 20s ./security/jwt ./x/tenant/session`
- `go test -race -timeout 60s ./security/jwt ./x/tenant/session`
- `go vet ./security/jwt ./x/tenant/session`

Docs Sync:
- Keep security docs aligned on the rule that stable `security` owns JWT verification primitives, while session lifecycle and revocation live in `x/tenant/session`.

Done Definition:
- Stable `security/jwt` no longer owns token revocation or version invalidation state.
- Session lifecycle semantics live in the owning tenant/session module.
- No residual references remain to removed stable lifecycle APIs.

Outcome:
- Completed.
- Removed token revocation and subject-version invalidation state from stable `security/jwt`; `JWTManager` now only signs, rotates keys, and verifies token semantics.
- Moved JWT-backed revocation/version state to `x/tenant/session` via `JWTStateStore`, which validates verified claims against revocation markers and subject versions.
- Updated security and tenant docs/manifests so session lifecycle ownership is explicitly documented under `x/tenant/session`.
