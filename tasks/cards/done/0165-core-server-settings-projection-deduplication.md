# Card 0165

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/config.go`
- `core/http_handler.go`
- `core/introspection.go`
- `core/app_test.go`
- `core/introspection_test.go`
Depends On:
- `0164-core-runtime-state-contract-collapse.md`

Goal:
- Deduplicate `core` server settings projection so `http.Server` preparation and
  runtime introspection derive from one canonical internal representation.

Problem:
- `AppConfig` is already the typed source of truth for server settings.
- `projectRuntimeSnapshot(cfg)` mirrors that config into `RuntimeSnapshot`.
- `ensureServerPrepared()` then mirrors the snapshot again into `http.Server`.
- This leaves `core` with two hand-maintained projection layers for the same
  server fields, increasing drift risk and making later cleanup harder.

Scope:
- Introduce one internal projection for server settings.
- Make server preparation and runtime introspection both derive from that same
  projection.
- Keep `AppConfig` as the caller-owned input contract.

Non-goals:
- Do not redesign the public `AppConfig` shape unless required by the chosen
  single projection.
- Do not add extra config wrappers or compatibility structs.
- Do not change routing or middleware behavior.

Files:
- `core/config.go`
- `core/http_handler.go`
- `core/introspection.go`
- `core/app_test.go`
- `core/introspection_test.go`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- No doc change unless the exported snapshot shape changes as part of the
  deduplication.

Done Definition:
- `core` has one canonical internal projection for server settings.
- `Prepare()` and runtime introspection no longer duplicate field mapping logic.
- Tests cover the normalized projection path instead of duplicated mappings.

Outcome:
- Replaced the duplicated `AppConfig -> RuntimeSnapshot -> http.Server`
  mapping chain with one internal `serverSettings` projection.
- Updated `RuntimeSnapshot()` and `ensureServerPrepared()` to derive from the
  same internal server-settings representation.
- Removed the old `projectRuntimeSnapshot(...)` helper so server settings are
  projected exactly once inside `core`.
