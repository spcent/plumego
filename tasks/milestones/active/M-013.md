# M-013: Migration Guides

- **Branch:** `milestone/M-013-migration-guides`
- **Depends on:** M-008
- **Parallel OK:** yes

---

## Goal

Publish `docs/migration/` with four concrete migration guides — from Gin, Echo, Chi, and stdlib `net/http` ServeMux — so Go developers already invested in another framework can evaluate plumego with a side-by-side comparison and runnable diff.

---

## Architecture Decisions

- Guides are Markdown files only; no code changes to plumego itself.
- Each guide must include: a feature comparison table, a before/after snippet for the top five patterns (routing, middleware, error response, request binding, response writing), and a migration checklist.
- The stdlib ServeMux guide explicitly addresses Go 1.22+ method+path syntax and explains what plumego adds beyond it (groups, named routes, middleware isolation, structured errors). This is the highest-priority guide for the Go community.
- Code snippets in guides are fenced Go blocks and must be syntactically valid (reviewable but not compiled in CI).
- Guides must not promise features that are Experimental; only Stable and Beta surfaces may appear as target patterns.
- A top-level `docs/migration/README.md` acts as the index.

---

## Context — Read Before Touching Code

1. `AGENTS.md`
2. `specs/task-routing.yaml`
3. `docs/ROADMAP.md`
4. `reference/standard-service/internal/app/routes.go`
5. `reference/standard-service/internal/app/app.go`
6. `contract/errors.go`
7. `router/router.go`

## Affected Modules

- **Primary:** `docs/migration/` (new directory)
- **Secondary:** `README.md`, `README_CN.md`, `docs/ROADMAP.md`

---

## Tasks

### Phase 1 — Orient (sequential)

1. Read every file in the **Context** section above.
2. Read `reference/standard-service/` fully to extract canonical patterns.
3. Identify the five canonical patterns to cover in each guide (routing, middleware, error, binding, response).

### Phase 2 — Implement (parallel)

- [ ] `docs/migration/README.md`: index with one-line description of each guide and a decision matrix (which framework you're coming from → which guide to read first).
- [ ] `docs/migration/from-stdlib.md`: migration from `net/http` ServeMux (Go 1.22+). Feature comparison table. Explicit callout of what plumego router adds beyond `{param}` syntax. Side-by-side for all five canonical patterns.
- [ ] `docs/migration/from-gin.md`: migration from Gin. Cover gin.Context vs. stdlib request/response, gin middleware vs. `func(http.Handler) http.Handler`, gin error handling vs. `contract.WriteError`.
- [ ] `docs/migration/from-echo.md`: migration from Echo. Cover Echo middleware, echo.HTTPError, echo.Bind.
- [ ] `docs/migration/from-chi.md`: migration from Chi. Cover chi.Router groups, chi middleware, chi URL params vs. plumego params.

### Phase 3 — Test (sequential)

6. Verify all Go code blocks in all guides are syntactically valid by a manual review pass.
7. Check all internal `docs/` cross-links resolve to existing files.

### Phase 4 — Validate and Ship (sequential)

8. Add a `## Migration Guides` section to `README.md` linking to `docs/migration/README.md`.
9. Mirror the same section in `README_CN.md`.
10. Run the **Acceptance Criteria** commands.
11. Commit: `docs(migration): add four migration guides [M-013]`
12. Final commit: `milestone(M-013): Migration Guides`
13. Push to `milestone/M-013-migration-guides`.

---

## Acceptance Criteria

```bash
go run ./internal/checks/module-manifests
gofmt -l .
```

```text
Manual check: docs/migration/ contains README.md and four guide files.
Manual check: all internal links in docs/migration/*.md resolve.
Manual check: README.md and README_CN.md link to docs/migration/README.md.
```

Expected: all commands exit 0; manual checks pass.

---

## Out of Scope

- Do not add new plumego features to satisfy guide patterns.
- Do not compile or test the guide code snippets in CI (they are illustrative).
- Do not write guides for frameworks not listed (Fiber, Gorilla, Beego, etc.) — follow-up work.
- Do not change any Go source files.

---

## Open Questions

(none at spec time)

---

## Done Definition

- [ ] `docs/migration/README.md` exists and lists all four guides.
- [ ] `docs/migration/from-stdlib.md` has a stdlib vs. plumego feature comparison table.
- [ ] `docs/migration/from-gin.md`, `from-echo.md`, `from-chi.md` exist with before/after snippets.
- [ ] `README.md` links to `docs/migration/README.md`.
- [ ] `README_CN.md` mirrors the same link.
- [ ] `gofmt -l .` produces no output.
- [ ] Branch `milestone/M-013-migration-guides` pushed.
- [ ] PR open, title `milestone(M-013): Migration Guides`.
