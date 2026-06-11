# Card 1527

Milestone: M-025
Recipe: (none — auth capability)
Context Package: use-cases/mini-saas-api
Priority: P1
State: active
Primary Module: use-cases/mini-saas-api
Owned Files: internal/domain/session, internal/handler/{auth,guard,me,errors}.go, internal/app wiring
Depends On: 1526

## Goal

Signup, login, and refresh work end to end: JWT access tokens (HS256 key
derived from APP_JWT_SECRET, persisted in store/kv), opaque rotated refresh
tokens with family invalidation on reuse, per-route RequireAuth guard, and
abuseguard on the auth endpoints.

## Scope

internal/domain/session (token lifecycle), handler auth/guard/me/errors,
app wiring (kv store, jwt manager, issuer, abuse guard), config APP_DATA_DIR.

## Non-goals

No tenant admin endpoints, no projects, no metrics (cards 1528-1530).
No cookie sessions, no logout-everywhere.

## Files

internal/domain/session/session.go
internal/handler/auth.go
internal/handler/guard.go
internal/app/app.go
internal/app/routes.go

## Acceptance Tests

internal/app/auth_flow_test.go: TestAcceptanceSignupLoginMeFlow
internal/app/auth_flow_test.go: TestAcceptanceRefreshRotation

## Tests

Session unit tests (rotation, reuse→family revoke, expiry, hash-only storage);
auth negative matrix (missing/garbage token, wrong password, unknown email,
duplicate email/slug, weak password); password hash never serialized.

## Docs Sync

env.example (APP_DATA_DIR added).

## Validation

cd use-cases/mini-saas-api && go vet ./... && go test -race -timeout 120s ./...
gofmt -l use-cases/mini-saas-api

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated.

## Outcome

session package owns refresh-token rotation with SHA-256-hashed storage and
family invalidation (reuse of a rotated token revokes the active successor).
JWT signing key derived deterministically from APP_JWT_SECRET and persisted
via store/kv (jwt.KeyStore satisfied by kv.KVStore). Tenant ID travels in the
JWT permissions claim ("tenant:<id>") and is lifted onto authn.Principal by
handler.RequireAuth. Roles re-derived from membership on refresh. Auth routes
wrapped in middleware/abuseguard. KV store and abuse guard released on
graceful shutdown via closeResources.
