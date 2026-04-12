# Card 0968: X Devtools Debug Alias Pruning

Priority: P1
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/devtools
Depends On: —

## Goal

Converge debug tooling on one canonical `/_debug/*` route vocabulary so
`x/devtools`, `cmd/plumego`, and dev-server docs stop maintaining two parallel
endpoint families for the same diagnostics.

## Problem

- `x/devtools/devtools.go` still mounts legacy top-level aliases `/_routes`,
  `/_config`, and `/_info` alongside the canonical `/_debug/*` endpoints.
- `cmd/plumego/commands/inspect.go` still calls those alias paths, while the
  dev-server analyzer and docs already teach `/_debug/routes.json`,
  `/_debug/config`, and `/_debug/metrics`.
- This splits the debug surface into two vocabularies, keeps alias-only tests
  alive, and makes route discovery less grep-friendly than it should be.

## Scope

- Define one canonical debug path set under `/_debug/*`, including a canonical
  replacement for `/_info` if that payload should remain exposed.
- Remove legacy top-level aliases from `x/devtools`.
- Update `cmd/plumego` inspect/analyzer consumers, tests, and docs to use only
  the canonical debug paths.

## Non-Goals

- Do not redesign debug payload schema beyond what is necessary to move to the
  canonical path set.
- Do not widen the card into pprof, env reload, or metrics behavior changes.
- Do not move devtools into `core`.

## Files

- `x/devtools/devtools.go`
- `x/devtools/devtools_test.go`
- `cmd/plumego/commands/inspect.go`
- `cmd/plumego/commands/*_test.go`
- `cmd/plumego/DEV_SERVER.md`

## Tests

```bash
rg -n '"/_routes"|"/_config"|"/_info"|/_routes|/_config|/_info' x/devtools cmd/plumego -g '*.*'
go test -timeout 20s ./x/devtools/... ./cmd/plumego/commands ./cmd/plumego/internal/devserver
go vet ./x/devtools/... ./cmd/plumego/commands ./cmd/plumego/internal/devserver
```

## Docs Sync

- `docs/modules/x-devtools/README.md` and `cmd/plumego/DEV_SERVER.md` for the
  canonical endpoint list

## Done Definition

- `x/devtools` exposes one canonical `/_debug/*` path family without top-level
  alias duplicates.
- `cmd/plumego` inspect/dev-server flows use the canonical debug paths only.
- Tests and docs no longer rely on `/_routes`, `/_config`, or `/_info`.

## Outcome

- Removed the legacy top-level `/_routes`, `/_config`, and `/_info` aliases
  from `x/devtools`.
- Added `DevToolsInfoPath` as the canonical `/_debug/info` endpoint and moved
  info payload tests to that path.
- Updated `cmd/plumego inspect` to use `/_debug/routes.json`,
  `/_debug/config`, and `/_debug/info`, and added CLI coverage that asserts the
  canonical endpoints are used.
- Updated the `x/devtools` module README to document the canonical `/_debug/*`
  endpoint family.

## Validation Run

```bash
rg -n '"/_routes"|"/_config"|"/_info"|/_routes|/_config|/_info' x/devtools cmd/plumego docs/modules/x-devtools -g '*.*'
go test -timeout 20s ./x/devtools/...
go vet ./x/devtools/...
cd cmd/plumego && go test -timeout 20s ./commands ./internal/devserver
cd cmd/plumego && go vet ./commands ./internal/devserver
```
