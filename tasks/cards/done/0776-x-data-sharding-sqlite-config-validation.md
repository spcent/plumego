# Card 0776

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/sharding/config
Owned Files:
- x/data/sharding/config/config.go
- x/data/sharding/config/config_test.go
- x/data/sharding/config/README.md
- docs/modules/x-data/README.md
Depends On:
- 0775-x-data-kvengine-close-timeout-flush

Goal:
Align sharding DatabaseConfig validation with driver-specific DSN construction and documented sqlite usage.

Scope:
- Validate sqlite3 configs with Database required and Host optional.
- Keep mysql/postgres host and database requirements explicit.
- Add focused config validation tests.

Non-goals:
- Do not add new database drivers.
- Do not change generated DSN strings except through validation rules.
- Do not change config watcher lifecycle.

Files:
- x/data/sharding/config/config.go
- x/data/sharding/config/config_test.go
- x/data/sharding/config/README.md
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding/config
- go test -race -timeout 60s ./x/data/sharding/config
- go vet ./x/data/sharding/config

Docs Sync:
- Update sharding config docs with driver-specific required fields.

Done Definition:
- Documented sqlite config validates without Host.
- mysql/postgres validation remains strict.
- Tests and docs cover the behavior.

Outcome:
- Made DatabaseConfig validation driver-specific for mysql, postgres, and sqlite3.
- Allowed documented sqlite3 configs to validate with only Database set.
- Rejected unsupported drivers when DSN is not provided.
- Updated sharding config docs and x/data dynamic config notes.

Validation:
- `go test -timeout 20s ./x/data/sharding/config`
- `go test -race -timeout 60s ./x/data/sharding/config`
- `go vet ./x/data/sharding/config`
