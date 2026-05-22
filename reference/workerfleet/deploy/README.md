# Workerfleet Kubernetes Deployment Assets

This directory contains reference manifests for running `reference/workerfleet`
inside one Kubernetes cluster.

Current assumptions:

- single cluster
- one active workerfleet replica by default
- Prometheus Operator style scraping through `ServiceMonitor`
- MongoDB, Feishu, and generic webhook credentials are injected externally

## Files

- `rbac.yaml`: `ServiceAccount`, `Role`, and `RoleBinding` for pod inventory sync
- `deployment.yaml`: `ConfigMap`, `Service`, and `Deployment`
- `servicemonitor.yaml`: Prometheus Operator scrape target for `/metrics`
- `prometheusrule.yaml`: baseline fleet health alert rules

## Required Secrets

Create the secret before applying `deployment.yaml` when you use these features:

- `workerfleet-secrets`
  - `mongo-uri`
  - `worker-auth-token`
  - `admin-auth-token`
  - `kube-bearer-token`
  - `feishu-webhook-url`
  - `webhook-url`

`worker-auth-token` and `admin-auth-token` are required by the reference
production profile. Other unused keys may be omitted when the corresponding
feature is disabled.

## Profile And Metrics Defaults

The reference deployment uses:

- `WORKERFLEET_PROFILE=prod`
- `WORKERFLEET_STORE_BACKEND=mongo`
- `WORKERFLEET_EXPERIMENTAL_METRICS_ENABLED=false`
- `WORKERFLEET_QUERY_AUTH_REQUIRED=true`
- `WORKERFLEET_KUBE_SYNC_ENABLED=true`
- `WORKERFLEET_STATUS_SWEEP_ENABLED=true`
- `WORKERFLEET_ALERT_EVALUATION_ENABLED=true`
- `WORKERFLEET_NOTIFICATION_ENABLED=false`

Those defaults keep stable fleet metrics on while leaving experimental pod,
`exec_plan_id`, and step-heavy metric families disabled unless explicitly
needed.
Notification delivery is disabled until at least one notifier secret is
configured and `WORKERFLEET_NOTIFICATION_ENABLED` is set to `true`.

## Apply Order

```bash
kubectl apply -f reference/workerfleet/deploy/rbac.yaml
kubectl apply -f reference/workerfleet/deploy/deployment.yaml
kubectl apply -f reference/workerfleet/deploy/servicemonitor.yaml
kubectl apply -f reference/workerfleet/deploy/prometheusrule.yaml
```

## Operational Notes

- The deployment intentionally starts with `replicas: 1`. The runtime loop layer
  already has a lease seam, but duplicate loop ownership is not implemented
  yet.
- `readinessProbe` uses `/readyz` so MongoDB and the runtime store must be
  initialized before the pod is considered ready.
- `livenessProbe` uses `/healthz` and avoids external dependency checks.
- The `Service` exposes the same port for API, health, and `/metrics`.
