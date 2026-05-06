# 0759 - Core Public Black-box Tests

State: active
Priority: P1
Primary module: core

## Goal

Add external package tests proving common core workflows work through public APIs only.

## Scope

- Add `package core_test` coverage for construction, route/middleware wiring, `ServeHTTP`, `Prepare`, `Server`, and `Shutdown`.
- Keep assertions on public behavior and exported types only.
- Cover route registration failure through public errors.

## Non-goals

- Do not inspect private `App` fields.
- Do not change lifecycle behavior.
- Do not add new public APIs.

## Files

- `core/core_public_test.go`
- `tasks/cards/active/0759-core-public-black-box-tests.md`

## Tests

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`

## Docs Sync

No docs update required; behavior is already documented.

## Done Definition

- External package tests cover the core happy path and a public failure path.
- Core normal and race tests pass.
