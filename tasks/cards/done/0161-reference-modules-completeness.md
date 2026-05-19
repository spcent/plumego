# Card 0161

Priority: P1
State: done
Primary Module: reference
Owned Files:
- `reference/*/go.mod`
- `reference/*/**/*.go`
- `reference/*/README.md`
- `README.md`
- `README_CN.md`
- `Makefile`
- `internal/checks/reference-layout/main.go`
- `internal/checks/checkutil/checkutil.go`
Depends On:
- Card 0160

Goal:
- Make every top-level `reference/*` directory a standalone, complete Go
  example module with its own `go.mod`, local imports, runnable README command,
  and explicit validation path.

Problem:
- Some `reference/*` examples are standalone modules, while others are still
  absorbed by the root module.
- This makes example ownership inconsistent and means adding `go.mod` files
  later can silently remove examples from root `go test ./...` coverage.
- Several examples import `github.com/spcent/plumego/internal/config`; once
  they become standalone modules, Go's `internal/` visibility rules reject
  that dependency.
- Existing docs still use root-relative commands such as
  `go run ./reference/standard-service`, which no longer match standalone
  module execution.

Scope:
- Add short-name `go.mod` files for every missing top-level reference example.
- Use `replace github.com/spcent/plumego => ../..` in every reference module.
- Rewrite each example's own internal imports to the short module path.
- Remove reference example dependencies on root `internal/*` packages.
- Update README commands to `cd reference/<example>` and `go run .`.
- Add a Makefile target that runs `go test ./...` in every reference module.
- Extend the reference layout check to enforce standalone reference modules.

Non-goals:
- Do not change Plumego stable public APIs.
- Do not add external dependencies.
- Do not promote root `internal/config` into a public package.
- Do not redesign the reference app layouts or route wiring.

Implementation Plan:
- Phase 1: add `go.mod` files and local import paths for simple examples.
- Phase 2: make app-structured examples standalone by localizing config
  loading and removing root `internal/config` imports.
- Phase 3: update docs and add automated reference module validation.
- Phase 4: run targeted reference module tests, boundary checks, and the new
  reference test target.

Tests:
- `make reference-test`
- `go run ./internal/checks/reference-layout`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/public-entrypoints-sync`

Docs Sync:
- Update root and reference README run commands so every standalone example is
  launched from its module directory.

Done Definition:
- Every top-level `reference/*` directory has a valid `go.mod`.
- Every reference module name equals its directory basename.
- Every reference module tests successfully with `go test ./...`.
- No reference Go file imports `github.com/spcent/plumego/reference/...`.
- No standalone reference module imports `github.com/spcent/plumego/internal/...`.
- The reference layout check fails if a future reference example misses
  `go.mod` or drifts back to long reference import paths.

Outcome:
- Added standalone `go.mod` files for the remaining reference examples.
- Rewrote reference-internal imports to use each example's short module path.
- Localized example config loading so standalone modules no longer import root
  `internal/config`.
- Updated README run commands to execute from each reference module directory.
- Added `reference-vet` and `reference-test` Makefile targets and wired both
  into `make gates`.
- Extended `reference-layout` to enforce standalone reference module shape,
  local Plumego replacement, and blocked root reference/internal imports.

Validation:
- `GOCACHE=/private/tmp/plumego-gocache make reference-test`
- `GOCACHE=/private/tmp/plumego-gocache make reference-vet`
- `GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/reference-layout`
- `GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/dependency-rules`
- `GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/agent-workflow`
- `GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/module-manifests`
- `GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/public-entrypoints-sync`
- `GOCACHE=/private/tmp/plumego-gocache go test ./internal/checks/checkutil`
- `GOCACHE=/private/tmp/plumego-gocache go test ./internal/checks/...`
- `GOCACHE=/private/tmp/plumego-gocache go test -timeout 20s ./...`
- `GOCACHE=/private/tmp/plumego-gocache go vet ./...`
