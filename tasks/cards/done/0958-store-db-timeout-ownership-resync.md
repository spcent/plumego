# Card 0958: Store DB Timeout Ownership Resync

Priority: P1
State: done
Primary Module: store

## Goal

Bring `store/db` back in line with the stable store boundary by removing or
isolating timeout-policy ownership from its public SQL helper surface.

## Problem

- `docs/modules/store/README.md` and the `store/db` package comment both say
  operation deadlines should be owned by callers through `context.Context`, and
  that generic timeout policy does not belong in stable `store/db`.
- The live API still exposes two timeout-owned stable hooks:
  - `Config.PingTimeout` / `DefaultConfig(...).PingTimeout`
  - `Ping(ctx, db, timeout)`
- `OpenWith(...)` currently creates its own `context.WithTimeout(...)` from
  `Config.PingTimeout`.
- `x/data/sharding/cluster.go` still consumes `store/db.Ping(ctx, shard, timeout)`,
  so the timeout-policy drift is not just dead surface; it is part of a current
  extension call path.

This leaves the stable DB helper layer split between the documented
caller-owned context model and a second stable timeout-policy API.

## Scope

- Decide one canonical timeout ownership path for stable `store/db`.
- Remove or de-emphasize stable timeout-policy surface that conflicts with the
  caller-owned context model.
- Migrate remaining callers to the chosen path in the same change.
- Resync `store/db` docs, examples, and tests with the final ownership model.

## Non-Goals

- Do not add retry loops, health payload models, or observability ownership to
  `store/db`.
- Do not widen `store/db` into a health manager or topology layer.
- Do not change unrelated query/transaction helpers.

## Files

- `store/db/sql.go`
- `store/db/sql_test.go`
- `docs/modules/store/README.md`
- any package docs/examples in `store/db`
- `x/data/sharding/cluster.go`
- any direct caller tests affected by the symbol/behavior change

## Tests

- `go test -timeout 20s ./store/db ./x/data/sharding/...`
- `go test -race -timeout 60s ./store/db ./x/data/sharding/...`
- `go vet ./store/db ./x/data/sharding/...`
- `go run ./internal/checks/dependency-rules`

## Docs Sync

- Keep stable store docs aligned with the resulting timeout ownership model.

## Done Definition

- Stable `store/db` no longer advertises a second timeout-policy surface that
  contradicts caller-owned context deadlines.
- All remaining callers use the canonical path consistently.
- Tests and docs describe one clear deadline model for DB ping/open behavior.

## Outcome

- Removed `Config.PingTimeout` and the implicit open-time ping path from
  `store/db`, leaving connection deadlines caller-owned through context.
- Changed `store/db.Ping` to consume the caller-provided context directly.
- Updated `x/data/sharding.ClusterDB.HealthCheck` to own timeout derivation via
  `context.WithTimeout(...)` before calling `store/db.Ping`.
- Resynced `store/db` comments/tests and stable store docs with the new
  deadline model.

## Validation Run

```bash
go test -timeout 20s ./store/db ./x/data/sharding/...
go test -race -timeout 60s ./store/db ./x/data/sharding/...
go vet ./store/db ./x/data/sharding/...
go run ./internal/checks/dependency-rules
go build ./...
```
