# Stable Root API Baseline

This directory records the current exported API surface for Plumego stable root
packages before a tagged v1 release.

The snapshots are baseline evidence, not a compatibility promise by themselves.
The compatibility promise starts only when a release tag and release notes make
that promise explicit.

## Stable Roots

| Root | Snapshot | Role |
| --- | --- | --- |
| `core` | `snapshots/core-head.snapshot` | App lifecycle, route attachment, middleware attachment, server assembly |
| `router` | `snapshots/router-head.snapshot` | Matching, params, groups, reverse routing |
| `contract` | `snapshots/contract-head.snapshot` | Response helpers, error model, context accessors, binding helpers |
| `middleware` | `snapshots/middleware-head.snapshot` | Transport-only HTTP middleware |
| `security` | `snapshots/security-head.snapshot` | Auth, headers, input safety, abuse guard primitives |
| `store` | `snapshots/store-head.snapshot` | Persistence primitives and base contracts |
| `health` | `snapshots/health-head.snapshot` | Health state and readiness models |
| `log` | `snapshots/log-head.snapshot` | Logging contracts and base implementations |
| `metrics` | `snapshots/metrics-head.snapshot` | Metrics contracts and base collectors |

## Regeneration

Use the existing API snapshot tool. Example:

```bash
go run ./internal/checks/extension-api-snapshot -module ./contract -out docs/stable-api/snapshots/contract-head.snapshot
```

For roots with subpackages, use `/...`:

```bash
go run ./internal/checks/extension-api-snapshot -module ./middleware/... -out docs/stable-api/snapshots/middleware-head.snapshot
```

## Freeze Rules

- Do not treat these files as permission to expand the stable surface.
- Signature changes, removals, and behavior changes in stable roots require
  explicit review against `docs/DEPRECATION.md`.
- Add behavior regression tests before changing high-risk paths in `core`,
  `router`, `contract`, `middleware`, or `security`.
- Keep stable roots free of `x/*` imports and new third-party dependencies.

## Required Checks

Before claiming stable-root freeze readiness:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go test -timeout 20s ./core/... ./router/... ./contract/... ./middleware/... ./security/... ./store/... ./health/... ./log/... ./metrics/...
go test -race -timeout 60s ./core/... ./router/... ./contract/... ./middleware/... ./security/... ./store/... ./health/... ./log/... ./metrics/...
go vet ./core/... ./router/... ./contract/... ./middleware/... ./security/... ./store/... ./health/... ./log/... ./metrics/...
```
