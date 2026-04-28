# 0675 - router static prefix normalization

State: active
Priority: P0
Primary module: `router`

## Goal

Make `Static` and `StaticFS` mount prefixes normalize consistently so trailing
slashes do not create internal double-slash route patterns.

## Scope

- Normalize static prefixes by trimming trailing slashes while preserving the
  root prefix.
- Cover `Static("/static/", ...)`, `StaticFS("/assets/", ...)`, and root static
  mount route listing behavior.
- Keep static serving as a small GET mount primitive.

## Non-goals

- Do not add SPA fallback, cache headers, MIME policy, or frontend asset policy.
- Do not change `Static` or `StaticFS` exported signatures.
- Do not change route matching precedence outside static mount registration.

## Files

- `router/static.go`
- `router/static_test.go`

## Tests

- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`

## Docs Sync

No docs sync expected; this preserves the documented static mount primitive.

## Done Definition

- Static mounts with trailing slash prefixes serve the same paths as normalized
  prefixes.
- Root static mounts register as `/*filepath`, not `//*filepath`.
- Router validation commands pass.

## Outcome

