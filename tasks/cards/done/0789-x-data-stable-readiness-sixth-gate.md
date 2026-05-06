---
id: 0789
title: x/data stable readiness sixth gate
status: done
priority: P3
primary_module: x/data
---

## Goal

Record the sixth x/data stable-readiness pass after cards 0781-0788 and keep
module docs/manifests synchronized with implemented behavior.

## Scope

- Sync `x/data/sharding/module.yaml` cross-shard policy wording.
- Update `docs/modules/x-data/README.md` stable blocker status.
- Run x/data package gates and boundary checks.
- Record remaining blockers without promoting the module to stable unless all
  release criteria are met.

## Non-goals

- Change support matrix status to stable without repo-wide release approval.
- Implement new behavior beyond documentation/manifest sync.
- Touch unrelated active cards.

## Files

- `x/data/sharding/module.yaml`
- `docs/modules/x-data/README.md`
- `tasks/cards/active/0789-x-data-stable-readiness-sixth-gate.md`

## Tests

- `go test -timeout 20s ./x/data/...`
- `go test -race -timeout 60s ./x/data/...`
- `go vet ./x/data/...`

## Docs Sync

Record exact commands run and remaining stable blockers.

## Done Definition

- x/data docs and manifest match the implemented behavior.
- x/data tests, race tests, vet, and boundary checks pass.
- Card is archived with validation evidence.

## Validation

- `go test -timeout 20s ./x/data/...`
- `go test -race -timeout 60s ./x/data/...`
- `go vet ./x/data/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
