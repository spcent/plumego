# Card 0756

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/config/config.go
- x/data/sharding/config/config_test.go
- x/data/sharding/config/watcher.go
- x/data/sharding/config/watcher_test.go
- docs/modules/x-data/README.md
Depends On:
- 0755-x-data-file-local-url-copy-durability

Goal:
Fail closed when env overrides make sharding config invalid and make generated DSNs safer.

Scope:
- Validate configuration after MergeWithEnv.
- Ensure watcher reload does not publish env-invalid configs.
- Escape DSN values through structured helpers.
- Keep existing development ssl defaults documented unless changed explicitly.

Non-goals:
- Do not introduce a new config format.
- Do not change runtime sharding behavior beyond invalid config rejection.
- Do not add non-stdlib dependencies.

Files:
- x/data/sharding/config/config.go
- x/data/sharding/config/config_test.go
- x/data/sharding/config/watcher.go
- x/data/sharding/config/watcher_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding/config
- go test -race -timeout 60s ./x/data/sharding/config
- go vet ./x/data/sharding/config

Docs Sync:
- Document env overrides are validated before publication.

Done Definition:
- Invalid env overrides return validation errors.
- Watcher callbacks are not called with invalid config.
- DSN construction safely handles reserved characters.
