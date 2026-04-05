# Card 0731

Priority: P3

Goal:
- Remove the `strings.TrimSpace` in `MustParam` so both `Param` and `MustParam`
  return the raw value, and document the emptiness check clearly.

Problem:

`context_core.go:343-349`:
```go
func (c *Ctx) MustParam(key string) (string, error) {
    val, ok := c.Param(key)
    if !ok || strings.TrimSpace(val) == "" {
        return "", errors.New("missing param: " + key)
    }
    return val, nil
}
```

`Ctx.Param` (lines 335-341) returns the raw value:
```go
func (c *Ctx) Param(key string) (string, bool) {
    val, ok := c.Params[key]
    return val, ok   // ← raw, not trimmed
}
```

The asymmetry means:
- `c.Param("key")` with value `"  "` → `("  ", true)` — non-empty, caller has to decide
- `c.MustParam("key")` with value `"  "` → `("", error)` — treated as missing

A route parameter containing only whitespace is highly unlikely in practice
(router patterns typically match non-empty segments), but the inconsistency is
still observable:

1. A caller who checks `val, ok := c.Param("key"); ok && val != ""` sees a
   different result than `val, err := c.MustParam("key")` for the same key.
2. `MustParam` silently returns a trimmed check result — the error says "missing
   param" but the param exists and is non-empty from the router's perspective.
3. `MustParam` returns `val` (the raw untrimmed value) but only if trimmed
   is non-empty — a caller gets the raw value but was checked against trimmed.

Fix: Remove `strings.TrimSpace`:
```go
func (c *Ctx) MustParam(key string) (string, error) {
    val, ok := c.Param(key)
    if !ok || val == "" {
        return "", errors.New("missing param: " + key)
    }
    return val, nil
}
```

Route parameters are set by the router from URL path segments, which cannot
contain whitespace in any standard routing scheme. The trimming is defensive
code that masks the raw value without benefit.

Non-goals:
- Do not add whitespace trimming to `Param`.
- Do not change the error message.

Files:
- `contract/context_core.go`

Tests:
- Add a test: `MustParam` with a key present but value `""` returns error.
- Add a test: `MustParam` with a key present and value `"abc"` returns `"abc"`.
- Confirm existing tests still pass.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `MustParam` does not call `strings.TrimSpace`.
- Both `Param` and `MustParam` operate on raw values.
- All tests pass.
