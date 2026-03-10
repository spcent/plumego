# Liveness Probes

> **Package**: `github.com/spcent/plumego/health` | **Purpose**: Detect whether process is alive

`LiveHandler` is intentionally simple: if the process can serve HTTP, liveness is `OK`.

## Canonical Usage

```go
app.Get("/health/live", health.LiveHandler().ServeHTTP)
```

Behavior:

- HTTP status: `200 OK`
- Body: `alive`

## When To Use

Use `/health/live` for orchestrators that should restart the container only when the process is no longer serving requests.

Do not place dependency checks (DB/cache/external APIs) in liveness. Those belong to readiness.

## Kubernetes Example

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 15
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 3
```

## Notes

- Keep liveness lightweight and constant-time.
- Dependency outages should fail readiness, not liveness.
- Pair with readiness endpoint for zero-downtime rollouts.
