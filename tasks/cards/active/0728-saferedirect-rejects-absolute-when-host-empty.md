# Card 0728

Priority: P2

Goal:
- Fix `validateRedirectURL` so that it does not reject all absolute same-origin
  URLs when `r.Host` is empty (which is common in unit tests).

Problem:

`context_response.go:111-117`:
```go
// Absolute URL: check that the host matches the request host
if r != nil && r.Host != "" {
    requestHost := r.Host
    if strings.EqualFold(parsed.Host, requestHost) {
        return nil
    }
}
return ErrUnsafeRedirect   // ← reached when r.Host == ""
```

When `r.Host` is empty, the guard `r.Host != ""` skips the host comparison and
falls through to `return ErrUnsafeRedirect`. This means any absolute URL is
rejected as unsafe when the request has no Host header — including valid
same-origin absolute URLs like `https://api.example.com/dashboard`.

This manifests in two situations:
1. **Unit tests**: `httptest.NewRequest` does not set `r.Host` by default.
   Calling `c.SafeRedirect(302, "https://example.com/path")` in a test always
   returns `ErrUnsafeRedirect`, making the method untestable without manually
   setting `r.Host`.
2. **Internal proxies**: Some internal proxying setups strip the Host header
   before forwarding, causing production redirects to fail unexpectedly.

Fix: when `r` is nil or `r.Host` is empty, fall back to allowing relative URLs
only (which is already handled by the `parsed.Scheme == "" && parsed.Host == ""`
check on line 107). For the absolute-URL path, if `r.Host` is empty, reject
rather than trying to compare — which is the current behaviour. The actual fix
needed is in tests and documentation:

**Option A**: Fall back to rejecting all absolute URLs when host is unknown
(current behaviour), but add a doc comment explaining this:
```go
// If the request has no Host header (r == nil or r.Host is empty),
// all absolute URLs are rejected as potentially unsafe.
// Use relative redirect paths in tests or ensure r.Host is set.
```

**Option B**: Accept absolute URLs when `r.Host` is empty if the URL host
matches an explicit allowlist provided via `RequestConfig`.

**Option C**: Extract the host from `r.URL` if `r.Host` is empty as a
fallback (`r.URL.Host` is populated by `httptest.NewRequest`).

Option C is preferred for testability:
```go
requestHost := r.Host
if requestHost == "" && r.URL != nil {
    requestHost = r.URL.Host
}
if requestHost != "" && strings.EqualFold(parsed.Host, requestHost) {
    return nil
}
```

Non-goals:
- Do not change the behavior for relative redirect paths.
- Do not add an open-redirect allowlist in this card.

Files:
- `contract/context_response.go`

Tests:
- Add a test: `SafeRedirect` with an `httptest.NewRequest` (no explicit Host)
  and a same-origin absolute URL must succeed.
- Add a test: cross-origin absolute URL still returns `ErrUnsafeRedirect`.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `SafeRedirect` is testable with `httptest.NewRequest` without manually setting
  `r.Host`.
- Cross-origin absolute URLs continue to be rejected.
- All tests pass.
