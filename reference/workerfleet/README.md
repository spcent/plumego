# Workerfleet Reference App

`reference/workerfleet` is the app-local worker monitoring service built on Plumego.

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
- [Technical Design](./docs/design/technical-design.md)
- [技术方案设计](./docs/design/technical-design.zh-CN.md)

Run locally:

- `cd reference/workerfleet && go run .`
- `cd reference/workerfleet && go build .`
- `cd reference/workerfleet && WORKERFLEET_HTTP_ADDR=:9090 go run .`

HTTP entrypoint configuration:

- `WORKERFLEET_HTTP_ADDR` controls the listen address, default `:8080`
- `WORKERFLEET_SHUTDOWN_TIMEOUT` controls graceful shutdown timeout, default `10s`
- `/metrics` is registered on the same HTTP server as the workerfleet API
- `/healthz` reports process liveness without external dependency checks
- `/readyz` reports readiness based on initialized runtime dependencies

Runtime loop configuration:

- `WORKERFLEET_KUBE_SYNC_ENABLED`, default `false`
- `WORKERFLEET_STATUS_SWEEP_ENABLED`, default `false`
- `WORKERFLEET_ALERT_EVALUATION_ENABLED`, default `false`
- `WORKERFLEET_NOTIFICATION_ENABLED`, default `false`
- `WORKERFLEET_KUBE_SYNC_INTERVAL`, default `30s`
- `WORKERFLEET_STATUS_SWEEP_INTERVAL`, default `30s`
- `WORKERFLEET_ALERT_EVALUATION_INTERVAL`, default `30s`
- `WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT`, default `5s`
- `WORKERFLEET_KUBE_API_HOST` optionally overrides in-cluster Kubernetes API discovery
- `WORKERFLEET_KUBE_BEARER_TOKEN` optionally overrides service account token discovery
- `WORKERFLEET_KUBE_NAMESPACE` controls the namespace watched by Kubernetes sync
- `WORKERFLEET_KUBE_LABEL_SELECTOR` limits watched worker pods
- `WORKERFLEET_KUBE_WORKER_CONTAINER` selects the worker container, default `worker`

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
- current state is stored separately from seven-day task, event, and alert history

Storage backend configuration:

- `WORKERFLEET_STORE_BACKEND=memory|mongo`
- `memory` is the default local backend
- `mongo` requires `WORKERFLEET_MONGO_URI` and `WORKERFLEET_MONGO_DATABASE`
- optional Mongo settings: `WORKERFLEET_MONGO_CONNECT_TIMEOUT`, `WORKERFLEET_MONGO_OPERATION_TIMEOUT`, `WORKERFLEET_MONGO_MAX_POOL_SIZE`
- `WORKERFLEET_RETENTION_DAYS` controls Mongo `expire_at` generation for task history, worker events, and alerts
