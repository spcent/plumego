# Card 1380

Milestone: agent-first-convergence
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/data
Owned Files:
- `x/data/**`
- `x/cache/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`
Depends On: Card 1379

Goal:
- Collapse cache topology and provider implementations into the data family path.

Scope:
- Move `x/cache` under `x/data`.
- Update imports, tests, taxonomy, package hotspots, dependency rules, and module manifests.
- Keep stable cache contracts in `store/cache`.

Non-goals:
- Do not change cache behavior.
- Do not move stable `store/cache`.
- Do not add provider dependencies.

Files:
- `x/data/**`
- `x/cache/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`

Tests:
- `go test -timeout 20s ./x/data/...`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Update `docs/modules/x-data/README.md` and `docs/modules/x-cache/README.md` if discovery or import paths change.

Done Definition:
- No Go file imports the old `github.com/spcent/plumego/x/cache` path.
- Data and cache tests pass under the data family.
- Agent workflow and dependency checks pass.

Outcome:
- Moved `x/cache` to `x/data/cache`.
- Updated module manifests, taxonomy, hotspots, dependency rules, maturity/docs/evidence references, and store/data guidance to the data family path.
- Verified `go test -timeout 20s ./x/data/...`, `go vet ./x/data/...`, `go run ./internal/checks/agent-workflow`, `go run ./internal/checks/dependency-rules`, and `go run ./internal/checks/module-manifests`.
