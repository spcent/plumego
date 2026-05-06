---
id: 0782
title: x/data sharding rule immutability and rewrite cache contract
status: active
priority: P2
primary_module: x/data/sharding
---

## Goal

Freeze the sharding rule mutation surface used by the registry and rewriter so
cached SQL rewrites cannot become stale after rule changes.

## Scope

- Prevent external mutation through registry accessors.
- Ensure registered rules are copied before storage.
- Ensure returned rule maps do not expose internal rule pointers.
- Add tests covering `SetActualTableName` after registration and cached rewrite
  behavior.

## Non-goals

- Remove existing exported fields.
- Introduce runtime hot-reload APIs for rule replacement.
- Change SQL rewriting semantics.

## Files

- `x/data/sharding/rule.go`
- `x/data/sharding/rewriter.go`
- `x/data/sharding/rule_test.go`
- `x/data/sharding/rewriter_test.go`

## Tests

- `go test -timeout 20s ./x/data/sharding`
- `go test -race -timeout 60s ./x/data/sharding`
- `go vet ./x/data/sharding`

## Docs Sync

Update `docs/modules/x-data/README.md` or `x/data/sharding/module.yaml` only if
the public ownership contract wording changes.

## Done Definition

- Registry-held rules cannot be mutated through caller-held pointers.
- `GetAll` returns independent rule values.
- Rewrite cache behavior remains deterministic after caller-side rule mutation.

