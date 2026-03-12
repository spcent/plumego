# AGENTS.md — plumego

Operational guide for AI coding agents in `spcent/plumego`.

## 1) Mission and Constraints

Plumego is a lightweight Go toolkit built on the standard library HTTP model.
Optimize for: clarity, explicit control flow, agent-friendly module ownership, small reversible changes.

Hard constraints:
- Preserve `net/http` compatibility
- Keep main module dependency-free (stdlib only)
- Do not blur module boundaries
- Do not introduce hidden global side effects
- Never log secrets

---

## 2) Canonical Style Authority

`docs/CANONICAL_STYLE_GUIDE.md` is the canonical style source for docs, examples, scaffolds, and AI-generated code.

Precedence when guidance overlaps:
1. Security and module-boundary constraints in this file
2. Coding style from `docs/CANONICAL_STYLE_GUIDE.md`
3. Existing local patterns in touched files

Canonical defaults:
- Handler shape: `func(http.ResponseWriter, *http.Request)`
- Explicit route registration: one method + path + handler per line
- Explicit JSON decode: `json.NewDecoder(r.Body).Decode(...)`
- Single error write path: `contract.WriteError` with structured error codes
- Constructor-based DI — no context service-locator pattern
- Middleware: `func(http.Handler) http.Handler`, transport-only responsibility

---

## 3) Module Boundaries (Strict)

| Module | Responsibility | Stability |
|---|---|---|
| `core/` | App lifecycle, options, startup/shutdown | High |
| `router/` | Routing, params, groups, reverse routing | High |
| `middleware/` | Cross-cutting HTTP middleware | High |
| `contract/` | Context, structured errors, response helpers | High |
| `security/` | JWT, input validation, headers, abuse guard | Critical |
| `health/` | Health models and readiness helpers | High |
| `log/` | Logging interfaces and base implementations | High |
| `metrics/` | Metrics interfaces and collectors | High |
| `store/` | Persistence abstractions | Medium |
| `x/tenant/` | Multi-tenancy extension boundary | Experimental |
| `x/*` | Optional or fast-evolving capability packs | Experimental |

Rules:
- Do not move routing behavior into `core`
- Do not put business logic in `middleware`
- Do not put tenant-aware logic in stable `middleware` or stable `store`
- Do not add new library code under broad legacy roots like `net/`, `utils/`, `validator/`, `tenant/`, `ai/`
- Changes in `core/`, `router/`, `middleware/`, `security/` require extra testing

Target layout:
- Stable library roots remain top-level: `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`
- Extension capability packs live under `x/*`
- `reference/` defines canonical app layout
- `examples/` are demos, not architectural authority

---

## 4) API and Change Rules

Do:
- Preserve stable public APIs unless explicitly asked otherwise
- Provide migration notes for unavoidable breaking changes
- Prefer standard library solutions over new dependencies
- Keep edits minimal and scoped to the task
- Add/update tests near changed behavior
- Update docs when changing API/config/security/default behavior

Do not:
- Introduce new handler styles or response/error helper families for one feature
- Hide DI through request context
- Inject business DTOs via middleware
- Add non-stdlib dependencies to the main module without strong reason

---

## 5) Security Rules

- Never log secrets, tokens, signatures, or private keys
- Fail closed on verification/authentication errors
- Use timing-safe comparison for secret checks
- Keep verification logic explicit and easy to review

---

## 6) Quality Gates (Required)

```bash
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

Extra checks by change type:
- Routing: static/param/group/reverse-routing tests
- Middleware: ordering and error-path tests
- Security: invalid token/signature negative tests
- Tenant: quota/policy/isolation tests
- Store: concurrent access and persistence correctness tests

---

## 7) Documentation Sync

Update when changing public APIs, env variables/defaults, security behavior, startup/shutdown semantics, or module boundaries.

Sync targets: `README.md`, `README_CN.md`, `AGENTS.md`, `CLAUDE.md`, `env.example`

---

## 8) Agent Workflow

1. Identify the target layer: stable root package or `x/*`
2. Read `specs/repo.yaml`, `specs/dependency-rules.yaml`, and the target `<module>/module.yaml`
3. Make minimal focused changes inside one primary module when possible
4. Add/update tests near changed behavior
5. Run quality gates (§6)
6. Sync docs if behavior or configuration changed
