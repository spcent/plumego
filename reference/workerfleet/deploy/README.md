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
  - `kube-bearer-token`
  - `worker-auth-token`
  - `feishu-webhook-url`
  - `webhook-url`

Unused keys may be omitted when the corresponding feature is disabled.

## Profile And Metrics Defaults

The reference deployment uses:

- `WORKERFLEET_PROFILE=prod`
- `WORKERFLEET_STORE_BACKEND=mongo`
- `WORKERFLEET_EXPERIMENTAL_METRICS_ENABLED=false`
- `WORKERFLEET_KUBE_SYNC_ENABLED=true`
- `WORKERFLEET_STATUS_SWEEP_ENABLED=true`
- `WORKERFLEET_ALERT_EVALUATION_ENABLED=true`

Those defaults keep stable fleet metrics on while leaving experimental
`exec_plan_id` and step-heavy metric families disabled unless explicitly needed.

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
