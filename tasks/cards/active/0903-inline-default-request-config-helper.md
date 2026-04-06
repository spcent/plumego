# Card 0903

Priority: P3
State: active
Primary Module: contract
Owned Files:
- `contract/context_core.go`
Depends On: —

Goal:
- Remove the private `defaultRequestConfig()` helper and inline its body into `newCtxWithConfig()`.

Problem:
`context_core.go` has three artefacts that all relate to the same default `RequestConfig`:

```go
// module-level var (line 195)
var defaultConfig = RequestConfig{ ... }

// private helper (line 208-211) — called exactly once
func defaultRequestConfig() *RequestConfig {
    cfg := defaultConfig
    return &cfg
}

// public accessor (line 204-206)
func DefaultConfig() RequestConfig {
    return defaultConfig
}
```

`defaultRequestConfig()` is a three-line function that copies `defaultConfig` into a pointer.
It is called in exactly one place — the nil-cfg branch of `newCtxWithConfig()` (line 219–221):

```go
if cfg == nil {
    cfg = defaultRequestConfig()
}
```

The indirection adds no clarity and splits a trivially simple operation across two functions.
The one-liner `cfg = &RequestConfig{}; *cfg = defaultConfig` (or equivalent `new`+copy) is
equally readable in place.

Scope:
- Delete `defaultRequestConfig()`.
- Inline the pointer-copy into the nil-cfg branch of `newCtxWithConfig()`.

Non-goals:
- Do not change `DefaultConfig()` or `defaultConfig`.
- Do not change `RequestConfig` fields.

Files:
- `contract/context_core.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `defaultRequestConfig()` no longer exists.
- `newCtxWithConfig()` inlines the pointer-copy inline.
- All tests pass.

Outcome:
