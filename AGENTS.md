# AGENTS.md — plumego

This document is the operational guide for AI coding agents working in `spcent/plumego`.

## 1) Mission and Constraints

Plumego is a lightweight Go toolkit built on the standard library HTTP model.
Agents must optimize for:

- Clarity and maintainability
- Explicit control/data flow
- Backward compatibility in stable APIs
- Small, reversible changes

Hard constraints:

- Preserve `net/http` compatibility
- Keep main module dependency-free (stdlib only)
- Do not blur module boundaries
- Do not introduce hidden global side effects
- Never log secrets

---

## 2) Canonical Style Authority (Mandatory)

`docs/CANONICAL_STYLE_GUIDE.md` is the canonical style source for docs, examples,
scaffolds, and AI-generated code.

When guidance overlaps, use this precedence:

1. Security and module-boundary constraints in this file
2. Coding style from `docs/CANONICAL_STYLE_GUIDE.md`
3. Existing local patterns in touched files

Canonical defaults:

- Standard-library handler shape: `func(http.ResponseWriter, *http.Request)`
- Explicit route registration: one method + path + handler per line
- Explicit JSON decode in handlers (`json.NewDecoder(r.Body).Decode(...)`)
- One stable error write path (`contract.WriteError` with structured error codes)
- Constructor-based dependency injection (no context service-locator pattern)
- Middleware as `func(http.Handler) http.Handler`, transport-only responsibility

Compatibility APIs may exist, but are not the default style for new docs/examples
unless a task explicitly asks for compatibility behavior.

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

- Do not move routing behavior into `core`.
- Do not put business logic in `middleware`.
- Do not put persistence/business logic in `utils`.
- Changes in `core/`, `router/`, `middleware/`, `security/` require extra testing.

---

## 4) API and Change Rules

Agents must:

- Preserve stable public APIs unless explicitly asked otherwise
- Provide migration notes for any unavoidable breaking behavior
- Prefer standard library solutions over new dependencies
- Keep edits minimal and scoped to the task

Avoid introducing:

- New handler styles
- New response/error helper families for one feature
- Hidden DI through request context
- Business DTO injection via middleware

---

## 5) Security Rules

- Never log secrets, tokens, signatures, or private keys
- Fail closed on verification/authentication errors
- Use timing-safe comparison for secret checks
- Keep verification logic explicit and easy to review

---

## 6) Quality Gates (Required)

Before finalizing changes, run:

```bash
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

## 7) Documentation Sync Rules

Update docs when changing:

- Public APIs
- Environment variables or defaults
- Security behavior
- Startup/shutdown semantics
- Module boundaries

Primary docs to keep in sync:

- `README.md`
- `README_CN.md`
- `AGENTS.md`
- `CLAUDE.md`
- `env.example`

---

## 8) Recommended Agent Workflow

1. Read `AGENTS.md`, `CLAUDE.md`, and relevant local files
2. Identify module boundary and risk level
3. Make minimal focused changes
4. Add/update tests near changed behavior
5. Run required quality gates
6. Update docs if behavior or configuration changed
7. Keep commit(s) small and reversible

---

Plumego values explicitness, restraint, and correctness over feature volume.
