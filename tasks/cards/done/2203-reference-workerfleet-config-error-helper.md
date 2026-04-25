# Card 2203: Reference Workerfleet Config Error Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: done
Primary Module: reference/workerfleet
Owned Files:
- `reference/workerfleet/internal/app/config_test.go`
Depends On: none

Goal:
Name the repeated workerfleet config error substring assertions.

Problem:
`config_test.go` repeats `err == nil || !strings.Contains(err.Error(), env)`
checks for invalid environment configuration. The tests are checking the same
contract, but each one spells out the nil and substring handling manually.

Scope:
- Add a local helper for asserting config errors mention the relevant env var.
- Use it in invalid runtime, webhook, and Mongo config tests.

Non-goals:
- Do not change config parsing or validation behavior.
- Do not add dependencies.

Files:
- `reference/workerfleet/internal/app/config_test.go`

Tests:
- from `reference/workerfleet`: `go test -race -timeout 60s ./internal/app/...`
- from `reference/workerfleet`: `go test -timeout 20s ./internal/app/...`
- from `reference/workerfleet`: `go vet ./internal/app/...`

Docs Sync:
No docs change required; this is test cleanup only.

Done Definition:
- Config validation tests use a named error assertion helper.
- The listed validation commands pass.

Outcome:
- Added `assertConfigErrorMentions` for workerfleet config validation tests.
- Validation passed for workerfleet app race tests, normal tests, and vet.
