# AGENTS.md — plumego

Operational guide for AI coding agents in `spcent/plumego`.

## 1) Mission and Constraints

Plumego is a lightweight Go toolkit built on the standard library HTTP model.
Optimize for: clarity, explicit control flow, backward compatibility, small reversible changes.

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
| `tenant/` | Multi-tenancy contracts and policy/quota primitives | High |
| `store/` | Persistence abstractions | Medium |
| `ai/` | AI gateway capabilities | Medium |
| `utils/` | Small shared helpers only | Low |

Rules:
- Do not move routing behavior into `core`
- Do not put business logic in `middleware`
- Do not put persistence/business logic in `utils`
- Changes in `core/`, `router/`, `middleware/`, `security/` require extra testing

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

1. Identify module boundary and risk level
2. Make minimal focused changes
3. Add/update tests near changed behavior
4. Run quality gates (§6)
5. Sync docs if behavior or configuration changed
