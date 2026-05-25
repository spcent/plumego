# Card 1518

Milestone: M-022
Recipe: specs/change-recipes/docs-and-config.yaml
Context Package: implementation
Priority: P2
State: done
Primary Module: reference/workerfleet
Owned Files:
- `reference/workerfleet/internal/app/service.go`
- `reference/workerfleet/README.md`
- `reference/workerfleet/AGENTS.md`
- `specs/deprecation-inventory.yaml`
Depends On: none

Goal:
- Reconcile workerfleet’s documented maturity and placeholder inventory with
  the current reality: core read/write flows exist, but several service methods
  still return `ErrNotImplemented` when optional backing dependencies are not
  wired in the current reference profile.

Scope:
- Document workerfleet as a WIP / partial reference rather than a richer
  production-style reference.
- Clarify in app service code what `ErrNotImplemented` means in the current
  profile.
- Update the deprecation inventory decisions so they describe partial
  dependency-gated behavior instead of broad unimplemented app claims.

Non-goals:
- Do not implement the remaining workerfleet service dependencies here.
- Do not change the workerfleet HTTP or storage behavior.

Files:
- `reference/workerfleet/internal/app/service.go`
- `reference/workerfleet/README.md`
- `reference/workerfleet/AGENTS.md`
- `specs/deprecation-inventory.yaml`

Acceptance Tests:
- `cd reference/workerfleet && go test -timeout 20s ./...`

Tests:
- `cd reference/workerfleet && go test -timeout 20s ./...`
- `go run ./internal/checks/deprecation-inventory -strict`

Docs Sync:
- `reference/workerfleet/README.md`
- `reference/workerfleet/AGENTS.md`

Validation:
- `cd reference/workerfleet && go test -timeout 20s ./...`
- `go run ./internal/checks/deprecation-inventory -strict`
- `gofmt -l .`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Repositioned workerfleet as a partial work-in-progress reference instead of a
  richer production-style template in both the local workerfleet README and its
  submodule AGENTS guide.
- Clarified in `internal/app/service.go` that `ErrNotImplemented` marks missing
  reference-profile dependency wiring, not wholesale absence of the service.
- Updated the workerfleet deprecation inventory decisions so they describe
  dependency-gated WIP behavior more accurately and attach that status to the
  app service file.
- Validation:
  - `cd reference/workerfleet && go test -timeout 20s ./...`
  - `go run ./internal/checks/deprecation-inventory -strict`
  - `gofmt -l .`
