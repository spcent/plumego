# Card 0083

Priority: P2

Goal:
- Add `SameSite http.SameSite` support to `Ctx.SetCookie` to close a
  security gap: the current signature omits SameSite, leaving browsers to
  apply their own default (which varies by browser and HTTPS context).

Problem:
- `contract/context_response.go:129-142`:
  ```go
  func (c *Ctx) SetCookie(name, value string, maxAge int, path, domain string,
      secure, httpOnly bool) {
      // ...
      http.SetCookie(c.W, &http.Cookie{
          // SameSite not set → browser default (typically SameSite=Lax in
          // modern browsers, but not guaranteed in all environments)
      })
  }
  ```
- OWASP recommends `SameSite=Strict` or `SameSite=Lax` for session cookies.
  Omitting the attribute is a CSRF risk in cross-origin contexts.
- `http.SameSite` constants are available since Go 1.11:
  `http.SameSiteDefaultMode`, `http.SameSiteStrictMode`, `http.SameSiteLaxMode`,
  `http.SameSiteNoneMode`.

Scope — two options (choose one):

**Option A (preferred): Accept `*http.Cookie` directly**
Replace the long parameter list with the standard library type, which already
has all cookie attributes including `SameSite`:
```go
// SetCookie adds a Set-Cookie header using the provided cookie configuration.
// Use http.Cookie{} to specify all attributes including SameSite.
func (c *Ctx) SetCookie(cookie *http.Cookie) {
    if cookie.Path == "" {
        cookie.Path = "/"
    }
    http.SetCookie(c.W, cookie)
}
```

**Option B (non-breaking): Add SameSite as last parameter**
```go
func (c *Ctx) SetCookie(name, value string, maxAge int, path, domain string,
    secure, httpOnly bool, sameSite http.SameSite) {
```
This is a breaking change to any callers that pass positional arguments.

Option A is preferred because it eliminates a long parameter list, is
forward-compatible (new http.Cookie fields become available automatically),
and matches the standard library directly.

Migration for Option A:
- Find all callers of `Ctx.SetCookie` and convert to `http.Cookie{}` struct
  literal. Run:
  ```bash
  grep -rn '\.SetCookie(' . --include='*.go'
  ```

Non-goals:
- Do not add a `DeleteCookie` helper in this card.
- Do not change `Ctx.Cookie` (getter).

Files:
- `contract/context_response.go`
- All callers of `Ctx.SetCookie` (grep first)

Tests:
- Add a test that verifies `SameSite` is propagated correctly via the
  `Set-Cookie` response header.
- `go test ./contract/...`
- `go vet ./...`
- `go build ./...`

Done Definition:
- `Ctx.SetCookie` accepts `*http.Cookie`, giving callers full attribute control.
- All existing callers migrated to the new signature.
- A test confirms `SameSite=Strict` in the Set-Cookie header when specified.
- All tests pass.
