# Working Rules for Agents

Read this file first. For exceptions and full context, see `AGENTS.md`.

---

## Non-Negotiables (must not violate — block merge if broken)

1. **No stable root imports `x/*`** — `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics` must never import anything under `x/`.
2. **No unapproved non-stdlib dependency** — `go.mod` must stay dependency-free. Any new import needs explicit human approval.
3. **Preserve `net/http` compatibility** — handler shape is always `func(http.ResponseWriter, *http.Request)`.
4. **Behavior changes require tests** — any change to observable behavior needs a focused test in the same PR.
5. **Fail closed on auth/verification/policy errors** — never fall back to permissive behavior on error.
6. **Never log secrets, tokens, signatures, or private keys**.
7. **No hidden globals, `init()` registration, or context service-locator patterns**.

---

## Preflight Checklist (fill before editing)

```
Owning module:
In-scope paths:
Out-of-scope paths:
Public API impact: none / yes (describe)
Security impact: none / yes (describe)
Validation plan: [exact commands]
```

---

## Gate Selection (use the lightest sufficient gate)

| Change type | Commands |
|---|---|
| Docs only | None required |
| Single stable-root file | `go test -race ./<module>/... && go vet ./<module>/...` |
| New middleware sub-package | Above + `go run ./internal/checks/dependency-rules` |
| `x/*` behavior change | `go test -race ./x/<family>/... && go run ./internal/checks/dependency-rules` |
| Cross-module or stable root | `make check-boundaries && make test-stable` |
| Public API change | `make gates-go` (full Go suite) |
| New dependency | **STOP** — requires human approval |

---

## Before You Edit

1. Read `<module>/module.yaml` — check `responsibilities`, `non_goals`, `stop_conditions`.
2. Read `specs/dependency-rules.yaml` entry for that module — check `allow` and `deny`.
3. For routing decisions read `specs/task-routing.yaml`.
4. For package ambiguity read `specs/package-hotspots.yaml`.

---

## Stop Conditions — switch to analysis-only mode when:

- Ownership of the change is unclear between two modules
- The change would add a new `go.mod` dependency
- A stable public API change was not explicitly requested
- The change would require a stable root to import `x/*`

When blocked, read `specs/stop-condition-handlers.yaml` for the exact resolution path.
