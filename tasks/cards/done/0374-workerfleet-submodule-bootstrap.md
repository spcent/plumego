# Card 0374

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet
Owned Files:
- `reference/workerfleet/go.mod`
- `reference/workerfleet/README.md`
- `reference/workerfleet/docs/storage.md`
Depends On:
- None
Blocked By:
- workerfleet MongoDB design review approval

Goal:
- Bootstrap `reference/workerfleet` as an independent Go submodule so MongoDB dependencies remain app-local and do not modify the repository root `go.mod`.
- Set the submodule path to `workerfleet` instead of `github.com/spcent/plumego/reference/workerfleet` so future repository extraction and migration do not depend on the current monorepo path.

Scope:
- Add `reference/workerfleet/go.mod` with a local replace back to the repository root.
- Introduce the MongoDB Go driver dependency only inside the `reference/workerfleet` submodule.
- Repoint app-local imports from `github.com/spcent/plumego/reference/workerfleet/...` to `workerfleet/...` within the submodule.
- Document the new submodule boundary and its validation expectations.

Non-goals:
- Do not add MongoDB repository implementations.
- Do not change workerfleet HTTP API contracts.
- Do not modify the repository root `go.mod`.

Files:
- `reference/workerfleet/go.mod`
- `reference/workerfleet/README.md`
- `reference/workerfleet/docs/storage.md`

Tests:
- `cd reference/workerfleet && go build ./...`
- `cd reference/workerfleet && go test ./...`

Docs Sync:
- Update `reference/workerfleet/README.md` to state that workerfleet is an app-local submodule with its own dependency surface.
- Update `reference/workerfleet/docs/storage.md` to describe the submodule-local persistence dependency policy.

Done Definition:
- `reference/workerfleet` builds and tests as an independent Go submodule.
- The repository root `go.mod` is unchanged.
- The submodule module path is `workerfleet`.
- App-local imports resolve as `workerfleet/...` rather than the monorepo-qualified path.
- MongoDB dependency scope is explicitly limited to `reference/workerfleet`.

Outcome:
- Added `reference/workerfleet/go.mod` with module path `workerfleet`, a local replace back to the repository root, and an app-local MongoDB driver dependency.
- Repointed workerfleet app-local imports from `github.com/spcent/plumego/reference/workerfleet/...` to `workerfleet/...`.
- Updated workerfleet docs to document the submodule boundary and submodule-local dependency policy.
- Validation run:
  - `cd reference/workerfleet && go build ./...`
  - `cd reference/workerfleet && go test ./...`
  - `go run ./internal/checks/dependency-rules`
- Repo-level checks still have unrelated pre-existing failures:
  - `go run ./internal/checks/module-manifests` fails on `/Users/bingrong.yan/projects/go/plumego/middleware/module.yaml: list "agent_hints" exceeds limit 3`
  - `go run ./internal/checks/reference-layout` fails on `website`
