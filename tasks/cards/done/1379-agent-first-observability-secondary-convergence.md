# Card 1379

Milestone: agent-first-convergence
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/observability
Owned Files:
- `x/observability/**`
- `x/ops/**`
- `x/devtools/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`
Depends On: Card 1378

Goal:
- Collapse protected operations and local dev diagnostics into the observability family path.

Scope:
- Move `x/ops` under `x/observability`.
- Move `x/devtools` under `x/observability`.
- Update imports, tests, taxonomy, package hotspots, dependency rules, and module manifests.

Non-goals:
- Do not change protected-route behavior or auth expectations.
- Do not move stable transport metrics middleware.
- Do not make devtools part of production bootstrap.

Files:
- `x/observability/**`
- `x/ops/**`
- `x/devtools/**`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`

Tests:
- `go test -timeout 20s ./x/observability/...`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Update `docs/modules/x-observability/README.md`, `docs/modules/x-ops/README.md`, and `docs/modules/x-devtools/README.md` if discovery or import paths change.

Done Definition:
- No Go file imports the old `github.com/spcent/plumego/x/ops` or `github.com/spcent/plumego/x/devtools` paths.
- Observability, ops, and devtools tests pass under the observability family.
- Agent workflow and dependency checks pass.

Outcome:
- Moved `x/ops` to `x/observability/ops` and `x/devtools` to `x/observability/devtools`.
- Updated CLI scaffold/devserver imports, reference docs, module manifests, taxonomy, hotspots, maturity signals, routing, and error-code paths to the observability family paths.
- Verified `go test -timeout 20s ./x/observability/... ./reference/with-ops/...`, `go vet ./x/observability/... ./reference/with-ops/...`, `go test -timeout 20s ./...` and `go vet ./...` in `cmd/plumego`, plus agent workflow, dependency, module manifest, and internal checks tests.
