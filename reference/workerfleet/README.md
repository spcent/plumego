# Workerfleet Reference App

`reference/workerfleet` is the app-local worker monitoring service built on Plumego.

Current documents:

- [API](./docs/api.md)
- [Storage](./docs/storage.md)

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
