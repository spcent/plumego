# 0738 - core Error Wrapping Regression

State: active
Priority: P1
Primary Module: core

## Goal

Lock the documented core error contract that lower-level causes remain wrapped
where available, without adding core sentinel errors.

## Scope

- Add regression coverage that TLS load failures wrap the underlying file
  system error.
- Keep error strings unchanged.
- Keep the no-new-core-sentinel decision unchanged.

## Non-goals

- Do not add exported error sentinel variables.
- Do not change `wrapCoreError` formatting.
- Do not widen public API.

## Files

- `core/lifecycle_test.go`

## Tests

- `go test -timeout 20s ./core/...`

## Docs Sync

Not required; existing core module docs already describe wrapped causes.

## Done Definition

- Tests prove callers can use `errors.Is` for wrapped lower-level causes where
  the producing package exports them.

