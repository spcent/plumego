# Plumego v1 GA Execution Plan (2026-03-10)

Status: In Progress  
Scope decision: Option A (`core/router/middleware/contract/security/store` GA; `tenant/*` + `net/mq/*` experimental)

## 1. Completed (This Round)

- API freeze scope documented:
  - `docs/other/V1_CORE_API_FREEZE.md`
  - `docs/other/V1_GA_PRODUCTION_SCOPE.md`
- Router reverse-routing + group boundary tests added.
- Middleware order + panic/error-path regression tests added.
- Security negative test matrix added (`jwt`, `headers`, `abuse`).
- README / README_CN canonical examples aligned (including multi-tenant snippets).
- Core entry docs aligned to current API:
  - `docs/getting-started.md`
  - `docs/modules/core/configuration-options.md`
  - `docs/modules/core/README.md`
  - `docs/modules/core/application.md`
  - `docs/modules/core/testing.md`
  - `docs/modules/core/lifecycle.md`
  - `docs/modules/core/components.md`
  - `docs/modules/core/dependency-injection.md`
- Full quality gates now pass:
  - `go test -timeout 20s ./...`
  - `go test -race -timeout 60s ./...`
  - `go vet ./...`
- Core/router named-route convenience APIs aligned and covered:
  - `core.App`: `GetNamed/PostNamed/PutNamed/DeleteNamed/PatchNamed/AnyNamed`
  - `router.Router`: `GetNamed/PostNamed/PutNamed/DeleteNamed/PatchNamed/AnyNamed`
  - Added regression tests in `core/routing_test.go` and `router/reverse_routing_group_test.go`
- Added strict core registration APIs with explicit errors:
  - `(*App).AddRoute(...) error`
  - `(*App).AddRouteWithName(...) error`
- Protocol middleware safety hardened:
  - nil registry passthrough behavior for `middleware/protocol`
  - request-body read error path handling
  - regression tests in `middleware/protocol/middleware_test.go`
- Canonical health/router/example docs corrected to compile-faithful API usage:
  - `health.ReadinessHandler(manager)` style
  - router reverse-routing `URL(name, "k", "v", ...)` style
- `scripts/check-doc-api-drift.sh` expanded:
  - detects old health API signatures
  - detects old router URL map signature pattern
  - scans `examples/docs/en` + `examples/docs/zh`
- `scripts/v1-release-readiness-check.sh` now passes end-to-end in current environment (root + submodules).

## 2. Current Risk Snapshot

### P0 (must finish before v1.0.0 final)

1. No new open P0 found in canonical v1 API/doc path after latest pass.

### P1 (should finish for production readiness)

1. Compile-faithful validation for canonical docs is still mostly pattern-based (`check-doc-api-drift.sh`), not full snippet compilation.
2. Historical/plan documents still contain removed APIs (intentional in some files), which can still confuse contributors if navigation discipline is not followed.

## 3. Drift Inventory (Top Files by stale API hit count)

From repository scan on 2026-03-10:

- `docs/modules/router/route-groups.md` (42)
- `docs/modules/router/middleware-binding.md` (29)
- `docs/modules/middleware/README.md` (17)
- `docs/legacy/topics/tenant/TENANT_PRODUCTION_PLAN.md` (13)
- `docs/legacy/topics/tenant/TENANT_IMPLEMENTATION_CHECKLIST.md` (9)
- `docs/modules/tenant/README.md` (8)
- `docs/modules/router/README.md` (8)
- `docs/modules/middleware/best-practices.md` (8)
- `docs/legacy/topics/tenant/TENANT_QUICKSTART.md` (7)
- `docs/modules/core/lifecycle.md` (5)

## 4. Execution Order (Remaining)

### Phase A: Canonical Routing/Middleware Docs

Target files:
- `docs/modules/router/route-groups.md`
- `docs/modules/router/middleware-binding.md`
- `docs/modules/router/README.md`
- `docs/modules/middleware/README.md`
- `docs/modules/middleware/best-practices.md`

Acceptance:
- No removed core middleware options in these files.
- All snippets use explicit `app.Use(...)` or `group.Use(...)`.
- Group examples use `app.Router().Group(...)` with current router API.

Status: Completed in current round.

### Phase B: Core Lifecycle and Component Docs

Target files:
- `docs/modules/core/lifecycle.md`
- `docs/modules/core/components.md`

Acceptance:
- Remove references to non-existent `app.Shutdown`, `app.Addr`, `app.Group`, `WithServer`, `WithDIContainer` in canonical sections.
- Keep lifecycle behavior aligned with current `Boot()` and signal handling model.

Status: Mostly completed for canonical paths; continue spot-fix as new drift appears.

### Phase C: Tenant Experimental Docs Cleanup

Target files:
- `docs/modules/tenant/*.md`
- `docs/legacy/topics/tenant/TENANT_QUICKSTART.md`

Acceptance:
- Replace removed core options with explicit tenant middleware chain on router groups.
- Add explicit “Experimental” compatibility banner in each entry doc.

Status: Partially completed; still ongoing for non-canonical planning docs.

### Phase D: Example Docs Cleanup (EN/ZH)

Target files:
- `examples/docs/en/*`
- `examples/docs/zh/*`

Acceptance:
- No removed API names in docs snippets.
- Keep examples aligned with README canonical style.

Status: Completed for high-traffic examples and now included in drift scan.

### Phase E: Canonical Snippet Compile Gate (New)

Target:
- Add a lightweight compile check for selected canonical docs snippets (`README*`, `docs/getting-started.md`, `docs/modules/core/README.md`, `docs/modules/health/*`, `docs/modules/router/README.md`, `docs/modules/router/basic-routing.md`, `docs/modules/router/middleware-binding.md`, `docs/modules/router/reverse-routing.md`, `docs/modules/router/route-groups.md`, `docs/modules/router/path-parameters.md`, `docs/modules/router/advanced-patterns.md`) plus high-traffic examples docs (`examples/docs/en|zh/guide.md`, `examples/docs/en/websocket/websocket.md`).

Acceptance:
- CI/local gate can catch signature drift beyond regex pattern checks.

Status: Completed in current round (`scripts/check-doc-snippets-compile.sh` integrated into release readiness gate, currently validating 24 `package main` snippets).

## 5. Release-Ready Exit Criteria

v1.0.0 final is eligible only when:

1. Quality gates are green:
   - `go test -timeout 20s ./...`
   - `go test -race -timeout 60s ./...`
   - `go vet ./...`
2. Canonical docs (`README`, `README_CN`, `docs/getting-started.md`, `docs/modules/core/*`) contain only compile-faithful API usage.
3. Removed API names do not appear in user-facing canonical docs except explicit compatibility notes.
4. Experimental modules (`tenant/*`, `net/mq/*`) are clearly marked as non-GA.

## 6. Suggested Working Strategy

- Continue in small, reversible PR-sized batches by phase.
- For each touched doc batch:
  1. update examples,
  2. run stale-API scan,
  3. run quality gates,
  4. sync README/README_CN if behavior or defaults changed.

## 7. Automation Entry

Use `scripts/v1-release-readiness-check.sh` as the single pre-release entrypoint.

- Default run includes:
  - root `gofmt` check
  - root `go test -timeout 20s ./...`
  - root `go vet ./...`
  - canonical doc drift check (`scripts/check-doc-api-drift.sh`)
  - canonical doc snippet compile check (`scripts/check-doc-snippets-compile.sh`)
  - submodule `gofmt` + `go test` + `go vet`
- Optional race run:
  - `RUN_RACE=1 bash scripts/v1-release-readiness-check.sh`
- Release operation reference:
  - `docs/other/V1_RELEASE_RUNBOOK.md`
