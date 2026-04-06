# Card 0905

Priority: P2
State: active
Primary Module: core
Owned Files:
- `core/config.go`
Depends On: —

Goal:
- Rename `AppConfig.EnableHTTP2` → `AppConfig.HTTP2Enabled` to match `RuntimeSnapshot.HTTP2Enabled`.
- Remove the `serverSettings` unexported struct and its two projection functions; project `AppConfig` directly to `RuntimeSnapshot`.

Problem:
`core/config.go` contains three structurally similar types:

```
AppConfig        → serverSettings → RuntimeSnapshot
  EnableHTTP2         HTTP2Enabled      HTTP2Enabled
```

Two issues compound here:

**1. Field name inconsistency.**
`AppConfig.EnableHTTP2` is the only field in the three types that is named with the
`Enable*` prefix instead of the descriptive adjective form. The same concept is called
`HTTP2Enabled` in `serverSettings` and `RuntimeSnapshot`. Callers who set `AppConfig`
and then introspect `RuntimeSnapshot` see two different names for the same toggle.

**2. Unnecessary middle layer.**
`serverSettings` is an unexported struct that duplicates all eight server-related fields
from `AppConfig`, with the sole purpose of feeding `runtimeSnapshot()`. Two boilerplate
copy functions (`projectServerSettings`, `runtimeSnapshot`) do nothing but shuttle the
same fields between three identical structs. The intermediate type adds indirection without
adding abstraction.

Fix:
- Rename `AppConfig.EnableHTTP2` → `AppConfig.HTTP2Enabled`.
- Delete `serverSettings` and `projectServerSettings`.
- Add `func (c AppConfig) RuntimeSnapshot(state PreparationState) RuntimeSnapshot` that
  projects `AppConfig` directly (inline the field copies).
- Update `core/http_handler.go` and `core/introspection.go` (or wherever `serverSettings`
  is used) to call `AppConfig.RuntimeSnapshot(...)` instead.
- Grep all callers: `grep -rn 'EnableHTTP2\|serverSettings\|projectServerSettings' . --include='*.go'`

Non-goals:
- Do not change any other `AppConfig` fields.
- Do not change `RuntimeSnapshot` or `PreparationState` types.
- Do not change TLS or Router config shapes.

Files:
- `core/config.go`
- `core/http_handler.go`
- `core/introspection.go`
- Any additional caller files found by grep

Tests:
- `go build ./...`
- `go test -timeout 20s ./core/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `AppConfig.EnableHTTP2` is renamed to `AppConfig.HTTP2Enabled`.
- `serverSettings` and `projectServerSettings` no longer exist.
- `AppConfig` has a `RuntimeSnapshot(PreparationState) RuntimeSnapshot` method.
- `grep -rn 'serverSettings\|projectServerSettings\|EnableHTTP2' . --include='*.go'` returns empty.
- All tests pass.

Outcome:
State: done
- `AppConfig.EnableHTTP2` renamed to `AppConfig.HTTP2Enabled`.
- `serverSettings` struct and `projectServerSettings()` function deleted.
- `AppConfig.runtimeSnapshot(PreparationState) RuntimeSnapshot` added, projecting directly.
- `core/http_handler.go` updated to use `cfg.*` directly instead of `settings.*`.
- `core/introspection.go` simplified to `return cfg.runtimeSnapshot(state)`.
- `grep -rn 'serverSettings\|projectServerSettings\|EnableHTTP2' . --include='*.go'` returns empty.
- `go test -timeout 20s ./...` passes.
