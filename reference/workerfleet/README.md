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
