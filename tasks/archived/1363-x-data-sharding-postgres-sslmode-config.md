---

State: done
id: 1363
title: x/data sharding postgres sslmode config
status: done
priority: P2
primary_module: x/data/sharding/config
---

## Goal

Make PostgreSQL SSL mode explicit in sharding database configuration instead of
hard-coding `sslmode=disable`.

## Scope

- Add `SSLMode` to `DatabaseConfig`.
- Validate known PostgreSQL sslmode values.
- Preserve current default for backward compatibility but document production
  override.
- Add DSN and validation tests.

## Non-goals

- Change default behavior in a breaking way.
- Add TLS certificate file configuration.
- Change MySQL or sqlite DSN behavior.

## Files

- `x/data/sharding/config/config.go`
- `x/data/sharding/config/config_test.go`
- `x/data/sharding/config/README.md`
- `docs/modules/x-data/README.md`

## Tests

- `go test -timeout 20s ./x/data/sharding/config`
- `go test -race -timeout 60s ./x/data/sharding/config`
- `go vet ./x/data/sharding/config`

## Docs Sync

Update config README and x/data docs with the explicit `ssl_mode` behavior.

## Done Definition

- PostgreSQL DSN uses configured sslmode when present.
- Invalid sslmode values are rejected by validation.
- Docs distinguish local default from production configuration.

## Validation

- `go test -timeout 20s ./x/data/sharding/config`
- `go test -race -timeout 60s ./x/data/sharding/config`
- `go vet ./x/data/sharding/config`
