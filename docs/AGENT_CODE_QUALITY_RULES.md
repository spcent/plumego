# Agent Code Quality Rules

This document turns Plumego's code-quality expectations into an agent-facing
contract. Use it with `AGENTS.md`, `docs/CODEX_WORKFLOW.md`,
`docs/AGENT_CONTEXT_BUDGET.md`, `docs/CANONICAL_STYLE_GUIDE.md`, and
`specs/agent-quality-rules.yaml`.

## 1. Rule Levels

### MUST

Violations block handoff or merge readiness unless a documented exception
exists.

- Stable roots must not import `x/*`.
- The main module must not add an unapproved non-stdlib dependency.
- `net/http` compatibility must be preserved.
- Behavior changes must include focused tests.
- Auth, verification, signature, and policy failures must fail closed.
- Secrets, tokens, signatures, and private keys must not be logged or returned.
- Hidden globals, `init()` registration, and context service-locator flows are forbidden.
- Exported symbol removal, rename, or behavior change must migrate every caller in the same change.
- HTTP JSON APIs must use the canonical `contract.WriteResponse` and `contract.WriteError` paths.

### SHOULD

- Keep each change centered on one primary module.
- Prefer standard-library shapes and existing local patterns.
- Place tests next to changed behavior.
- Keep route, handler, and middleware wiring grep-friendly.
- Document implemented behavior only.
- Run module checks first, boundary checks second, and repo-wide gates last.
- Load the smallest safe context package.
- Summarize validation output compactly.

### MAY

- Extract a small helper when local repetition is clear and stable.
- Skip Go gates for docs-only changes.
- Keep a compatibility layer only with an explicit migration window.
- Split broad work into an analysis pass or task card before coding.

## 2. Preflight

Before editing, answer:

```text
Context package:
Owning module:
Target module.yaml read:
In-scope paths:
Out-of-scope paths:
Public API impact: none / yes
Dependency impact: none / yes
Behavior impact: none / yes
Security impact: none / yes
Docs impact: none / yes
Validation plan:
```

Stop in analysis-only mode when ownership, dependency impact, stable public API
impact, or cross-module boundaries are unclear.

Use `docs/AGENT_CONTEXT_BUDGET.md` to choose `startup`, `implementation`,
`review`, or `control-plane` before reading beyond the selected
`specs/task-routing.yaml` `start_with` files.

## 3. Module Quality Focus

### `core`

- Keep lifecycle order explicit.
- Return route registration errors.
- Attach middleware visibly.
- Keep shutdown controllable.
- Do not turn `core` into a feature catalog.

### `router`

- Cover static routes, params, groups, method matching, and reverse routing.
- Do not add response writing, auth policy, or business validation.
- Keep matching behavior grep-friendly.

### `contract`

- Preserve error-code and response-shape stability.
- Keep `contract` limited to transport primitives.
- Use `With{Type}` and `{Type}FromContext` accessor pairs.
- Use unexported zero-value struct context keys inlined at call sites.
- Do not add new response helper families or specific `NewXxxError` helpers.
- Remove deprecated symbols after caller migration.

### `middleware`

- Call `next` exactly once.
- Test ordering and error paths.
- Keep behavior transport-only.
- Do not inject business services, build business DTOs, or branch on domain policy.
- Fallible constructors should return `(..., error)`.

### `security`

- Use timing-safe comparison for secrets, tokens, and signatures.
- Fail closed on invalid or unverifiable input.
- Add negative tests for invalid token and signature paths.
- Keep sensitive values out of logs and responses.
- Do not hide security state behind context service-location.

### `store`

- Protect concurrent reads and writes.
- Propagate context cancellation.
- Keep persistence behavior testable.
- Do not import `core`, `router`, `middleware`, `security`, or `x/*`.

### `x/tenant`

- Preserve tenant isolation.
- Cover quota, policy, and session lifecycle with negative tests.
- Keep tenant behavior out of stable `middleware` and `store`.

## 4. Anti-Patterns

Do not introduce:

- Stable package imports of `x/*`
- New dependencies without approval
- Context-based service lookup
- Package-level mutable registries
- `init()` side-effect registration
- Route auto-discovery or reflection-based route wiring
- Middleware that builds business DTOs or injects business services
- Ad hoc JSON response helpers
- Per-feature error envelopes
- New handler signatures
- New panic-only constructors for fallible behavior
- Generic `utils` packages for new library code
- Compatibility wrappers without a removal plan
- Deprecated symbols left after caller migration

## 5. Review Output Contract

Use findings-first output for reviews:

```text
Findings:
- [P0/P1/P2] Title
  File:
  Problem:
  Impact:
  Minimal fix:
  Missing test:

Boundary status:
Validation status:
Docs sync:
Residual risk:
```

Review priorities:

1. Boundary violations
2. Stable-root imports of `x/*`
3. `net/http` compatibility risk
4. Hidden globals or context service-location
5. Fail-open security behavior
6. Response or error path drift
7. Missing behavior tests
8. Docs, config, or example drift

## 6. Gate Profiles

### docs-only

Check accuracy, links, terminology, and authority order. Go gates are not
required unless code, config, generated data, or runnable examples changed.

### single module behavior

```bash
go test -race -timeout 60s ./<module>/...
go test -timeout 20s ./<module>/...
go vet ./<module>/...
gofmt -w <changed-go-files>
```

### stable root change

```bash
go test -race -timeout 60s ./<module>/...
go test -timeout 20s ./<module>/...
go vet ./<module>/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
```

### router, middleware, or security

```bash
go test -race -timeout 60s ./<module>/...
go test -timeout 20s ./<module>/...
go vet ./<module>/...
go run ./internal/checks/dependency-rules
```

Also confirm:

- `router`: static, param, group, method, and reverse-route tests
- `middleware`: ordering, error-path, panic, timeout, and cancel tests
- `security`: invalid-token, invalid-signature, fail-closed, and timing-safe tests

### cross-module, public API, or release-relevant

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go run ./internal/checks/extension-maturity
go run ./internal/checks/extension-beta-evidence
go run ./internal/checks/deprecation-inventory -strict
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

## 7. Handoff Contract

Every implementation handoff should include:

```text
Context package:
Owning module:
Files changed:
Boundary impact:
API/dependency impact:
Tests added or updated:
Validation commands run:
Docs sync impact:
Residual risk:
```
