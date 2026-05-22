# Workerfleet Reference App

`reference/workerfleet` is the app-local worker monitoring service built on Plumego.
It is not the canonical v1 reference application and is not a reusable stable
surface for Plumego applications. Use `reference/standard-service` for v1
application structure, route wiring, and stable-root-only onboarding.

Submodule boundary:

- `reference/workerfleet` builds as the `workerfleet` Go module
- app-local packages import each other as `workerfleet/...`
- Plumego root packages are consumed through `replace github.com/spcent/plumego => ../..`
- MongoDB dependencies stay scoped to this submodule and do not modify the repository root `go.mod`

Current documents:

- [API](./docs/api.md)
- [Storage](./docs/storage.md)
- [Metrics](./docs/metrics.md)
- [Case And Step Metrics Design](./docs/case-step-metrics.md)
- [Grafana Dashboard Plan](./docs/grafana.md)
- [Kubernetes Deployment Assets](./deploy/README.md)
- [Technical Design](./docs/design/technical-design.md)
- [技术方案设计](./docs/design/technical-design.zh-CN.md)

Run locally:

- `cd reference/workerfleet && go run .`
- `cd reference/workerfleet && go build .`
- `cd reference/workerfleet && WORKERFLEET_HTTP_ADDR=:9090 go run .`
- [env.example](./env.example) includes the current workerfleet env surface for local bootstrapping

HTTP entrypoint configuration:

- `WORKERFLEET_HTTP_ADDR` controls the listen address, default `:8080`
- `WORKERFLEET_SHUTDOWN_TIMEOUT` controls graceful shutdown timeout, default `10s`
- `/metrics` is registered on the same HTTP server as the workerfleet API
- `/healthz` reports process liveness without external dependency checks
- `/readyz` reports readiness based on initialized runtime dependencies

Worker ingress authentication:

- `WORKERFLEET_WORKER_AUTH_TOKEN` enables Bearer-token auth for `POST /v1/workers/register` and `POST /v1/workers/heartbeat`
- when unset, worker ingress auth is disabled for local development
- when set, missing, malformed, or invalid credentials fail closed with `401`
- `WORKERFLEET_PROFILE=prod` requires `WORKERFLEET_WORKER_AUTH_TOKEN` at startup

Query API authentication:

- `WORKERFLEET_ADMIN_AUTH_TOKEN` enables Bearer-token auth for query endpoints such as `/v1/workers`, `/v1/tasks/:task_id`, `/v1/fleet/summary`, and `/v1/alerts`
- `WORKERFLEET_QUERY_AUTH_REQUIRED=true` requires `WORKERFLEET_ADMIN_AUTH_TOKEN` even outside the production profile
- worker ingress auth and query API auth use separate tokens so worker pods cannot query fleet state unless explicitly granted the admin token
- `/healthz`, `/readyz`, and `/metrics` are not covered by query API auth in this reference app
- `WORKERFLEET_PROFILE=prod` requires `WORKERFLEET_ADMIN_AUTH_TOKEN` at startup

Runtime loop configuration:

- `WORKERFLEET_PROFILE=dev|prod`, default `dev`
- `WORKERFLEET_KUBE_SYNC_ENABLED`, default `false`
- `WORKERFLEET_STATUS_SWEEP_ENABLED`, default `false`
- `WORKERFLEET_ALERT_EVALUATION_ENABLED`, default `false`
- `WORKERFLEET_NOTIFICATION_ENABLED`, default `false`
- `WORKERFLEET_KUBE_SYNC_INTERVAL`, default `30s`
- `WORKERFLEET_STATUS_SWEEP_INTERVAL`, default `30s`
- `WORKERFLEET_ALERT_EVALUATION_INTERVAL`, default `30s`
- `WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT`, default `5s`
- `WORKERFLEET_EXPERIMENTAL_METRICS_ENABLED`, default `true` in `dev` and `false` in `prod`
- when `WORKERFLEET_NOTIFICATION_ENABLED=true`, at least one notifier URL must be configured
- `WORKERFLEET_KUBE_API_HOST` optionally overrides in-cluster Kubernetes API discovery
- `WORKERFLEET_KUBE_BEARER_TOKEN` optionally overrides service account token discovery
- `WORKERFLEET_KUBE_NAMESPACE` controls the namespace watched by Kubernetes sync
- `WORKERFLEET_KUBE_LABEL_SELECTOR` limits watched worker pods
- `WORKERFLEET_KUBE_WORKER_CONTAINER` selects the worker container, default `worker`
- runtime loop errors are exported as `workerfleet_runtime_errors_total`
- each runtime loop runs with built-in non-overlap, a default `25s` iteration timeout, a `5s` initial failure backoff, and a `1m` max failure backoff
- a no-op lease seam is already present in the loop scheduler so future multi-replica ownership can be added without rewriting loop bodies
- when Kubernetes sync, status sweep, or alert evaluation are enabled, keep the current deployment at `replicas: 1` until distributed lease ownership is implemented

Status and alert policy configuration:

- `WORKERFLEET_PROFILE=dev` uses the app-local development policy defaults
- `WORKERFLEET_PROFILE=prod` switches to more conservative production defaults for heartbeat, offline, stage-stuck, and restart-burst thresholds
- `WORKERFLEET_STATUS_STALE_AFTER` overrides worker heartbeat staleness
- `WORKERFLEET_STATUS_OFFLINE_AFTER` overrides worker offline detection
- `WORKERFLEET_STATUS_STAGE_STUCK_AFTER` overrides worker stage-stuck detection
- `WORKERFLEET_STATUS_RESTART_BURST_THRESHOLD` overrides restart-burst status evaluation
- `WORKERFLEET_ALERT_STAGE_STUCK_AFTER` overrides alert-side stage-stuck firing
- `WORKERFLEET_ALERT_RESTART_BURST_THRESHOLD` overrides alert-side restart-burst firing
- policy values are validated at startup; unsafe low thresholds fail closed before the HTTP server is exposed
- `WORKERFLEET_EXPERIMENTAL_METRICS_ENABLED` overrides the profile default for pod, exec-plan, and case-step experimental metric families

Single-cluster Kubernetes assumptions:

- the service reads pods from one Kubernetes cluster
- pod discovery is namespace-scoped
- label selectors are used to limit the watched worker pod set
- the service expects a service account token or an explicit bearer token

Minimum RBAC expectations:

- `get`, `list`, and `watch` on pods in the target namespace

Current monitoring model:

- one pod maps to one worker
- workers report the full active-task set on each heartbeat
- active cases can include `exec_plan_id` and `current_step`
- case step timeline and exec-plan drilldown APIs expose case-level detail for Grafana links
- current state is stored separately from seven-day task, event, and alert history

Storage backend configuration:

- `WORKERFLEET_STORE_BACKEND=memory|mongo`
- `memory` is the default local backend
- `mongo` requires `WORKERFLEET_MONGO_URI` and `WORKERFLEET_MONGO_DATABASE`
- optional Mongo settings: `WORKERFLEET_MONGO_CONNECT_TIMEOUT`, `WORKERFLEET_MONGO_OPERATION_TIMEOUT`, `WORKERFLEET_MONGO_MAX_POOL_SIZE`
- `WORKERFLEET_RETENTION_DAYS` controls Mongo `expire_at` generation for task history, worker events, and alerts; values must be greater than zero and no more than 106751 days
