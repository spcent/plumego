# AGENTS.md — plumego

Operational guide for AI coding agents working in `github.com/spcent/plumego`.

## 1. Goal

Plumego is an agent-first Go toolkit built on the standard-library HTTP model.
Optimize for:

- clear module ownership
- explicit control flow
- small reversible changes
- minimal search radius
- one canonical implementation path

## 2. Non-Negotiables

- Preserve `net/http` compatibility.
- Keep the main module dependency-free (stdlib only) unless explicitly approved.
- Do not blur stable-module boundaries.
- Do not introduce hidden globals, `init()` registration, or context service-locator patterns.
- Never log secrets, tokens, signatures, or private keys.
- Fail closed on auth, verification, and policy errors.
- Use timing-safe comparison for secret checks.
- Context accessor pairs follow `With{Type}` + `{Type}FromContext` order; context
  key types are unexported zero-value structs inlined at the call site — no package-level variable.
- `contract` contains transport primitives only. Tracing infrastructure, session
  lifecycle management, and metric collection do not belong in `contract`.
- Deprecated symbols must be removed in the same PR that replaces their last
  caller. Do not leave dead wrappers behind.
- One canonical success-response path per layer; one canonical error-construction
  path per layer. Do not add per-feature response helpers or per-scenario error constructors.

## 3. Read Order

Use this default path before making changes:

1. `docs/CANONICAL_STYLE_GUIDE.md`
2. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
3. `specs/repo.yaml`
4. `specs/agent-entrypoints.yaml`
5. `specs/dependency-rules.yaml`
6. `specs/ownership.yaml`
7. target `<module>/module.yaml`
8. `reference/standard-service`

For staged future work and sequencing, also read:

- `docs/ROADMAP.md`
- `specs/change-recipes/*`
- `tasks/cards/*`

When guidance overlaps, follow:

1. security and boundary rules in this file
2. `docs/CANONICAL_STYLE_GUIDE.md`
3. machine-readable repo specs
4. existing local patterns in touched files

## 4. Canonical Defaults

- Handler shape: `func(http.ResponseWriter, *http.Request)`
- Route wiring: one method + path + handler per line
- JSON decode: `json.NewDecoder(r.Body).Decode(...)`
- Error write path: `contract.WriteError` with structured error codes
- DI: constructor-based and explicit in route wiring
- Middleware: `func(http.Handler) http.Handler`, transport-only responsibility
- Reference app: `reference/standard-service` is the only canonical application layout

## 5. Agent Navigation Rules

Five rules determine where to work. Read this before scanning the entrypoints list.

| Intent | Destination |
|---|---|
| Change kernel, lifecycle, route structure, transport contracts, transport middleware, auth primitives, storage primitives | stable root |
| Change product capability, business feature, protocol adaptation, extension behavior | `x/*` |
| Change application wiring, bootstrap, DI, route registration | `reference/standard-service` or `internal/scaffold` |
| Change architecture rules, boundary definitions, quality gates | `specs/` |
| Change execution plan, work items, task sequencing | `tasks/cards/` |

**Stable roots (9):** `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`

**x/* primary families (10):** `x/tenant`, `x/messaging`, `x/gateway`, `x/rest`, `x/websocket`, `x/frontend`, `x/observability`, `x/files`, `x/data`, `x/ai`

Always start at a primary family, not a subordinate (`x/mq`, `x/pubsub`, `x/ops`, `x/cache`, `x/devtools`, etc.).
See `specs/agent-entrypoints.yaml` for the full routing table and detailed entrypoints.

## 6. Module Boundaries

Stable library roots:

- `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`

Extension roots:

- `x/*` for optional or fast-evolving capabilities

Hard rules:

- Stable roots must not depend on `x/*`.
- `core` is the app kernel, not a feature catalog.
- `router` owns matching, params, groups, and reverse routing.
- `middleware` stays transport-only; never hide business DTO assembly or service injection there.
- `contract` owns transport contracts and response/error helpers, not protocol gateway families,
  observability infrastructure, or session lifecycle management.
  Tracing subsystems (Tracer, Span, Collector, Sampler) belong in `x/observability`.
  Session management (SessionStore, SessionValidator, RefreshManager) belongs in `x/tenant`.
- `health` owns models/readiness helpers, not HTTP handler ownership.
- Tenant-aware logic belongs in `x/tenant`, not stable `middleware` or stable `store`.
- Stable `middleware` must not grow tenant resolution, tenant policy, or tenant quota behavior.
- Stable `store` must not grow tenant-aware adapters or tenant-specific storage policy.
- New library code must live under a stable root or `x/*`; avoid broad legacy roots such as `net`, `utils`, `validator`, `tenant`, `ai`, `rest`, `pubsub`.

Task entrypoint defaults:

- HTTP endpoint work: start with the style guide, `reference/standard-service/internal/app/routes.go`, `contract`, and `router`.
- Middleware work: start with `middleware/module.yaml` and `docs/modules/middleware/README.md`.
- Security work: start with `security/module.yaml` and `docs/modules/security/README.md`.
- Store work: start with `store/module.yaml` and `docs/modules/store/README.md`.
- Gateway or edge transport work: start with `x/gateway` (includes service discovery and IPC).
- Resource API standardization: start with `x/rest`.
- Messaging work: start with `x/messaging` (not `x/mq` or `x/pubsub` directly).
- Tenant work: start with `x/tenant` and `docs/architecture/X_TENANT_BLUEPRINT.md`.
- WebSocket transport work: start with `x/websocket`.
- File upload/download/storage work: start with `x/files`.
- Admin or observability surfaces: start with `x/observability` (includes ops and devtools), not `health`.
- AI capability work: start with `x/ai`.
- Data topology work (sharding, rw-split, cache): start with `x/data`.

## 7. Change Rules

- Keep changes minimal and scoped to one primary module when possible.
- Preserve stable public APIs unless explicitly asked to change them.
- If a breaking change is unavoidable, add migration notes.
- Prefer standard-library solutions over new abstractions.
- Add or update tests next to changed behavior.
- Do not invent one-off handler styles, response envelopes, or helper families for a single feature.

## 8. Validation Order

Default order:

1. Run the target module tests from `specs/ownership.yaml` or `<module>/module.yaml`.
2. Run boundary and manifest checks.
3. Run repo-wide gates before final handoff.

Required repo-wide gates:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

Extra checks by change type:

- Routing: static, param, group, and reverse-routing coverage
- Middleware: ordering and error-path coverage
- Security: invalid token/signature negative tests
- Tenant: quota, policy, and isolation tests
- Store: concurrent access and persistence correctness tests

`specs/check-baseline/` contains temporary migration debt baselines. Reduce them; do not expand them casually.
If a baseline file is empty, treat the file itself as migration debt and prefer removing the placeholder once the corresponding check can tolerate a missing baseline file.

## 9. Docs Sync

Update these when behavior, public API, config, security semantics, lifecycle behavior, or boundaries change:

- `README.md`
- `README_CN.md`
- `AGENTS.md`
- `CLAUDE.md`
- `docs/ROADMAP.md`
- `env.example`

## 10. Working Loop

1. Identify the target layer: stable root or `x/*`.
2. Read the canonical sources in Section 3.
3. If a matching repo-native task card exists, follow it.
4. Confirm the owning module manifest and validation commands.
5. Make the smallest coherent change.
6. Add or update focused tests.
7. Run validation in the order from Section 7.
8. Sync docs only for implemented behavior changes.
