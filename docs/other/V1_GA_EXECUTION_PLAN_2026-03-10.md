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

## 2. Current Risk Snapshot

### P0 (must finish before v1.0.0 final)

1. Documentation drift still exists in module docs/examples with removed APIs (`WithRecommendedMiddleware`, `WithTenantMiddleware`, `app.Shutdown`, etc.).
2. Canonical API story is still fragmented across router/middleware/tenant docs.

### P1 (should finish for production readiness)

1. Cross-doc compatibility annotation is inconsistent (stable vs experimental module boundaries).
2. Example catalog contains historical snippets that are not compile-faithful.

## 3. Drift Inventory (Top Files by stale API hit count)

From repository scan on 2026-03-10:

- `docs/modules/router/route-groups.md` (42)
- `docs/modules/router/middleware-binding.md` (29)
- `docs/modules/middleware/README.md` (17)
- `docs/tenant/TENANT_PRODUCTION_PLAN.md` (13)
- `docs/tenant/TENANT_IMPLEMENTATION_CHECKLIST.md` (9)
- `docs/modules/tenant/README.md` (8)
- `docs/modules/router/README.md` (8)
- `docs/modules/middleware/best-practices.md` (8)
- `docs/tenant/TENANT_QUICKSTART.md` (7)
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

### Phase B: Core Lifecycle and Component Docs

Target files:
- `docs/modules/core/lifecycle.md`
- `docs/modules/core/components.md`

Acceptance:
- Remove references to non-existent `app.Shutdown`, `app.Addr`, `app.Group`, `WithServer`, `WithDIContainer` in canonical sections.
- Keep lifecycle behavior aligned with current `Boot()` and signal handling model.

### Phase C: Tenant Experimental Docs Cleanup

Target files:
- `docs/modules/tenant/*.md`
- `docs/tenant/TENANT_QUICKSTART.md`

Acceptance:
- Replace removed core options with explicit tenant middleware chain on router groups.
- Add explicit “Experimental” compatibility banner in each entry doc.

### Phase D: Example Docs Cleanup (EN/ZH)

Target files:
- `examples/docs/en/*`
- `examples/docs/zh/*`

Acceptance:
- No removed API names in docs snippets.
- Keep examples aligned with README canonical style.

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
  - submodule `gofmt` + `go test` + `go vet`
- Optional race run:
  - `RUN_RACE=1 bash scripts/v1-release-readiness-check.sh`
- Release operation reference:
  - `docs/other/V1_RELEASE_RUNBOOK.md`
