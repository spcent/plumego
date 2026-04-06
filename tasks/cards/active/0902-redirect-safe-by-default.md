# Card 0902

Priority: P1
State: active
Primary Module: contract
Owned Files:
- `contract/context_response.go`
Depends On: —

Goal:
- Make the default `Redirect()` method safe (same-origin only) and rename the unrestricted variant to `UnsafeRedirect()`.

Problem:
`Ctx` currently exposes two redirect methods:

```go
// Redirect — accepts ANY URL including external hosts
func (c *Ctx) Redirect(status int, location string) error { ... }

// SafeRedirect — restricts to relative paths and same-origin URLs
func (c *Ctx) SafeRedirect(status int, location string) error { ... }
```

The naming is a security footgun. Callers who reach for `Redirect()` — the obvious,
unsuffixed name — get the unrestricted variant that allows open-redirect attacks.
The secure alternative is buried under a `Safe` prefix that callers must know to seek out.
Good API design puts the safe behaviour on the natural name and makes unsafe behaviour
explicit and deliberate.

Fix:
- Rename `SafeRedirect()` → `Redirect()` (becomes the canonical call).
- Rename `Redirect()` → `UnsafeRedirect()` (explicit opt-in for cross-origin redirects).
- Update `ErrUnsafeRedirect` doc comment if needed (the name stays correct).
- Grep all callers: `grep -rn 'Redirect\b\|SafeRedirect' . --include='*.go'`

Scope:
- `contract/context_response.go`: swap method names and update comments.
- All caller files: migrate `SafeRedirect(...)` → `Redirect(...)`, `Redirect(...)` → `UnsafeRedirect(...)`.

Non-goals:
- Do not change redirect validation logic (`validateRedirectURL`).
- Do not add new redirect policies.

Files:
- `contract/context_response.go`
- Any caller files found by grep

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `Redirect()` performs same-origin validation (former `SafeRedirect` behaviour).
- `UnsafeRedirect()` accepts any URL (former `Redirect` behaviour).
- `SafeRedirect` no longer exists in the codebase.
- `grep -rn 'SafeRedirect' . --include='*.go'` returns empty.
- All tests pass.

Outcome:
