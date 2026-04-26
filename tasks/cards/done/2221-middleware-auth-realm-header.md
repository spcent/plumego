# Card 2221

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files: middleware/auth/auth.go; middleware/auth/auth_test.go; tasks/cards/done/2221-middleware-auth-realm-header.md
Depends On: 2220

Goal:
Build `WWW-Authenticate` realm headers safely instead of concatenating unescaped user-supplied realm strings.

Scope:
- Quote and escape realm values as HTTP auth quoted strings.
- Strip control characters from realm values before writing response headers.
- Preserve existing 401/403 auth error mapping and optional custom error handler behavior.
- Add tests for quoted/backslash realm values and control-character stripping.

Non-goals:
- Do not implement token verification logic in middleware.
- Do not change `security/authn` interfaces.
- Do not change authorization policy semantics.

Files:
- `middleware/auth/auth.go`
- `middleware/auth/auth_test.go`

Tests:
- `go test -timeout 20s ./middleware/auth`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; behavior remains the same API with safer header construction.

Done Definition:
- Realm headers remain valid for special characters and cannot include raw CR/LF/control bytes.
- Auth error mapping tests still pass.
- Targeted middleware tests and vet pass.

Outcome:
- Added a small `Bearer realm=` builder that strips controls and escapes quote/backslash characters.
- Added tests for quoted/backslash realm values and CR/LF/tab stripping.
- Validation run: `go test -timeout 20s ./middleware/auth`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
