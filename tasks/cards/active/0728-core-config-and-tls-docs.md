# 0728 - core Config And TLS Docs

State: active
Priority: P2
Primary Module: core

## Goal

Clarify core stable configuration semantics for zero-value server hardening
fields and TLS ownership.

## Scope

- Document zero timeout, zero `MaxHeaderBytes`, and non-negative validation
  behavior for `AppConfig`.
- Document that core only loads configured certificate/key material and callers
  own advanced TLS policy by adjusting the prepared `*http.Server`.
- Keep runtime behavior unchanged.

## Non-goals

- Do not add TLS policy defaults.
- Do not change HTTP/2 behavior.
- Do not add new config fields.

## Files

- `docs/modules/core/README.md`

## Tests

- `go run ./internal/checks/module-manifests`

## Docs Sync

Required in `docs/modules/core/README.md`.

## Done Definition

- Core module docs clearly describe zero-value config semantics and TLS
  ownership without promising unimplemented behavior.

