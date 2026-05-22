# AGENTS.md - reference/workerfleet

Operational guide for agents working in `reference/workerfleet`.

Workerfleet is an app-local monitoring reference service built on Plumego. It is
not the canonical Plumego application template and it is not a stable root
package. Use this file together with the repository root `AGENTS.md`; when the
two conflict, the root security and boundary rules still win.

## 1. Current Shape

- Module path: `workerfleet`.
- Location: `reference/workerfleet`.
- Plumego dependency: `replace github.com/spcent/plumego => ../..`.
- Canonical app reference remains `reference/standard-service`; workerfleet is a
  richer production-style reference for worker monitoring.
- MongoDB and workerfleet-specific Kubernetes, alerting, notifier, and metrics
  dependencies stay inside this submodule.

Do not move workerfleet behavior into Plumego stable roots. Do not add
workerfleet-specific labels, routes, storage models, or business policy to root
packages such as `core`, `router`, `contract`, `middleware`, `security`,
`store`, `health`, `log`, or `metrics`.

## 2. Package Ownership

- `main.go`: thin process entrypoint only. Load config, construct the app, run
  it. Keep detailed middleware, route, loop, and shutdown orchestration out of
  `main.go`.
- `internal/app`: application bootstrap, server config, explicit dependency
  wiring, runtime loop orchestration, graceful shutdown, and app-level service
  command/result/view types.
- `internal/handler`: HTTP request parsing, route registration, worker ingress
  auth enforcement, HTTP DTOs, app-to-HTTP DTO mapping, and
  `contract.WriteResponse` / `contract.WriteError` calls.
- `internal/domain`: worker status policy, worker/task reconciliation, pod
  reconciliation, alert rules, domain events, and domain-owned store ports.
- `internal/platform/store`: app-local storage interfaces, query filters, and
  persistence-facing records.
- `internal/platform/store/memory`: local in-memory store implementation.
- `internal/platform/store/mongo`: MongoDB implementation, indexes, document
  mapping, and operation timeout handling.
- `internal/platform/kube`: Kubernetes client, pod mapping, and inventory sync.
- `internal/platform/metrics`: workerfleet Prometheus collector and observer.
- `internal/platform/notifier`: Feishu and generic webhook sinks.

## 3. Boundary Rules

- `internal/app` must not import `internal/handler`. Preserve the current
  registrar injection boundary.
- `internal/handler` may import `internal/app` for app-level service interfaces
  and app-owned command/result types.
- Domain packages must not depend on HTTP, MongoDB, Kubernetes, Prometheus, or
  notifier packages.
- Store interfaces accept `context.Context`; do not add storage methods that use
  `context.Background()` instead of the caller context.
- HTTP DTOs stay in `internal/handler`; app service methods expose app-level
  commands/results from `internal/app`.
- Workerfleet-specific dependencies must not be added to the repository root
  `go.mod`.

## 4. HTTP And Route Style

- Handler shape remains `func(http.ResponseWriter, *http.Request)`.
- JSON decode remains `json.NewDecoder(r.Body).Decode(...)`.
- Success responses use `contract.WriteResponse`.
- Error responses use `contract.WriteError`.
- Do not introduce a second response envelope or app-local generic error helper
  family.
- Route registration stays static and grep-friendly. The current route table in
  `internal/handler/routes.go` is acceptable; keep each route visible as one
  method/path/handler entry.
- If route wiring needs more cross-package dependencies, prefer adding named
  fields to a dependency struct over growing long positional function
  signatures.

## 5. Security Rules

- Never log or echo secrets, bearer tokens, webhook headers, signatures, or
  private keys.
- `WORKERFLEET_WORKER_AUTH_TOKEN` controls worker ingress auth. When it is set,
  `POST /v1/workers/register` and `POST /v1/workers/heartbeat` must fail closed
  on missing, malformed, or invalid credentials.
- `WORKERFLEET_PROFILE=prod` must fail startup when
  `WORKERFLEET_WORKER_AUTH_TOKEN` is missing.
- Worker token checks must use timing-safe comparison. Do not replace the
  current fixed-length digest plus constant-time compare pattern with plain
  string equality.
- Query endpoints are not covered by worker ingress auth unless a task
  explicitly changes the public access model and updates docs/tests.
- Kubernetes bearer tokens and notifier webhook headers must stay out of error
  messages and metric labels.

## 6. Runtime And Observability

- Runtime loops live in `internal/app`, not in handlers or `main.go`.
- Kubernetes sync, status sweep, alert evaluation, and notification delivery
  errors must be observed through the runtime error observer; do not silently
  discard loop errors.
- Metrics labels must stay low-cardinality. Do not add worker IDs, task IDs,
  case IDs, pod UIDs, raw error messages, tokens, or secrets as default
  Prometheus labels.
- Shutdown ordering matters: stop background loops, shut down HTTP serving, then
  close runtime resources.
- Readiness should check initialized runtime dependencies; liveness should avoid
  external dependency checks.

## 7. Change Rules

- Keep one primary module per change when possible.
- Do not refactor stable Plumego roots while working on workerfleet unless the
  task explicitly asks for it.
- Preserve existing route paths, JSON field names, status codes, and error
  envelopes unless the task explicitly changes the API contract.
- When changing exported symbols or interfaces, enumerate callers with `rg`
  first and update all call sites in the same change.
- If you add or rename environment variables, update docs in the same change.
- Do not document planned behavior as implemented behavior.

## 8. Docs Sync

Update the relevant docs when behavior, API payloads, environment variables,
security semantics, runtime lifecycle, metrics, or storage behavior changes:

- `README.md`
- `docs/api.md`
- `docs/design/technical-design.md`
- `docs/design/technical-design.zh-CN.md`
- `docs/metrics.md`
- `docs/alerts.md`
- `env.example`
- repository root `env.example` when environment variables are added or renamed

Keep the English and Chinese technical design docs aligned.

## 9. Validation

Use the lightest sufficient checks for the change:

- Handler or route changes:
  - `cd reference/workerfleet && go test -timeout 20s ./internal/handler/...`
- App, config, runtime loop, or bootstrap changes:
  - `cd reference/workerfleet && go test -timeout 20s ./internal/app/...`
- Domain rule changes:
  - `cd reference/workerfleet && go test -timeout 20s ./internal/domain/...`
- Store changes:
  - `cd reference/workerfleet && go test -timeout 20s ./internal/platform/store/...`
  - include memory or Mongo subpackage tests when touched
- Cross-package or behavior-bearing changes:
  - `cd reference/workerfleet && go test -timeout 20s ./...`
- Before handoff:
  - `git diff --check`

For docs-only changes, `git diff --check` is the minimum. Run workerfleet tests
when docs are coupled to behavior you changed in the same task.

## 10. Review Focus

When reviewing workerfleet changes, prioritize:

- app/handler reverse dependency regressions.
- request context not reaching store operations.
- background loop errors being swallowed.
- worker ingress auth fail-open behavior.
- secret leakage in logs, errors, docs examples, or metric labels.
- high-cardinality metric labels.
- HTTP response shape drift.
- technical design drift from implementation.
