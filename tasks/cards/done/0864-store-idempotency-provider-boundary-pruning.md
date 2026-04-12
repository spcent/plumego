# Card 0864: Store Idempotency Provider Boundary Pruning

Priority: P1
State: done
Primary Module: store

## Goal

Keep stable `store/idempotency` limited to small reusable idempotency primitives. Move durable provider implementations, SQL dialect policy, table schema policy, and feature-specific dedupe ownership out of stable `store`.

## Problem

`store/idempotency` currently owns more than a primitive contract:

- `SQLStore`
- `NewSQLStore`
- `SQLConfig`
- `DefaultSQLConfig`
- `Dialect`
- SQL schema/table/cleanup behavior
- SQL duplicate-key detection policy
- `KVStore`
- `NewKVStore`
- `KVConfig`
- `DefaultKVConfig`

The stable `store` manifest says business-level idempotency semantics belong in applications or `x/*`. The only non-test runtime consumer of the SQL adapter is currently `x/mq/dedupe_sql.go`, which means stable `store` is carrying a provider implementation for an extension feature.

## Scope

- Preserve a small stable idempotency contract only if it remains genuinely reusable.
- Move SQL-backed idempotency storage out of stable `store`.
- Prefer `x/data/idempotency` for a generic durable provider unless implementation inspection shows the SQL provider is only MQ-specific; in that case, place it under `x/mq`.
- Move SQL dialect/config/table policy with the provider.
- Move KV-backed idempotency adapter out of stable `store` if it is not needed as a primitive.
- Update `x/mq` dedupe code to depend on the new provider owner.
- Remove deprecated wrappers and old exported provider symbols in the same change.
- Update docs and module manifests to describe the narrowed stable store surface.

## Non-Goals

- Do not remove MQ SQL dedupe capability.
- Do not add an external dependency.
- Do not turn stable `store` into a tenant-aware or feature-aware persistence layer.
- Do not keep deprecated provider aliases after callers are migrated.

## Expected Files

- `store/idempotency/*`
- `x/mq/dedupe_sql.go`
- `x/mq/*_test.go`
- `x/data/**` if the generic provider target is used
- `docs/modules/store/README.md`
- `store/module.yaml`
- `x/data/module.yaml`
- `x/mq/module.yaml`

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./store/idempotency ./x/data/... ./x/mq/...
go test -race -timeout 60s ./store/idempotency ./x/mq/...
go vet ./store/idempotency ./x/data/... ./x/mq/...
```

Then run the required repo-wide gates before committing.

## Done Definition

- Stable `store/idempotency` no longer exports durable SQL/KV provider implementations unless a primitive-level adapter is explicitly justified in the final diff.
- SQL dialect and table policy live with the extension provider owner.
- `x/mq` SQL dedupe capability still works.
- Removed exported provider symbols have zero residual references.
- Focused gates and repo-wide gates pass.

## Outcome

- Stable `store/idempotency` now contains only the reusable idempotency primitive contract: errors, `Status`, `Record`, and `Store`.
- Moved SQL and KV provider implementations, SQL dialects, table policy, duplicate detection, and provider tests into `x/data/idempotency`.
- Added `x/data/idempotency/module.yaml` and documented the provider ownership split in store and x/data docs.
- Updated `x/mq` SQL dedupe to depend on `x/data/idempotency` while preserving MQ dedupe capability.
- Removed stable provider exports (`SQLStore`, `NewSQLStore`, `SQLConfig`, `DefaultSQLConfig`, `Dialect`, `KVStore`, `NewKVStore`, `KVConfig`, `DefaultKVConfig`) from `store/idempotency`.
- Validation passed: focused store/x-data/x-mq test, race, and vet gates; dependency rules, agent workflow, module manifests, reference layout; repo-wide `go test`, `go vet`, and `go test -race`.
