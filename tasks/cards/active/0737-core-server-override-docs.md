# 0737 - core Server Override Docs

State: active
Priority: P2
Primary Module: core

## Goal

Clarify ownership boundaries when callers mutate the `*http.Server` returned by
`App.Server()`.

## Scope

- Document that caller-side server field overrides are caller-owned.
- Call out fields that can bypass core guarantees, such as `Handler`,
  `ConnState`, `TLSConfig`, and `TLSNextProto`.
- Keep runtime behavior unchanged.

## Non-goals

- Do not wrap or restrict the returned `*http.Server`.
- Do not add new server config fields.
- Do not change TLS or HTTP/2 behavior.

## Files

- `README.md`
- `README_CN.md`
- `docs/modules/core/README.md`

## Tests

- `go run ./internal/checks/reference-layout`

## Docs Sync

Required in English and Chinese top-level docs plus core module docs.

## Done Definition

- Docs make clear which behavior remains core-owned and which behavior becomes
  caller-owned after direct server mutation.

