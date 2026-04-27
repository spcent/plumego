# Card 0329: Admin And Cluster HTTP Surface Convergence

Priority: P1
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: multi-module admin transport
Depends On: —

## Goal

Make the scheduler admin endpoints and pubsub cluster/export endpoints follow a
single, explicit HTTP response policy instead of mixing `contract` writes with
hand-rolled JSON, plain-text method errors, and ad hoc status handling.

## Problem

- `x/scheduler/admin_http.go` defines its own `writeJSON` helper, mixes
  `contract.WriteError`, `http.NotFound`, and manual JSON responses, and leaves
  the admin surface with multiple response shapes.
- `x/pubsub/distributed.go` writes cluster health, heartbeat, and sync payloads
  with raw `json.NewEncoder(w).Encode(...)`, while nearby deny paths already use
  `contract.WriteError`.
- `x/pubsub/prometheus.go` returns plain-text `http.Error` for method mismatch
  even though the package already has a defined HTTP surface and the rest of the
  repo prefers explicit error construction.
- These files all represent app-facing or operator-facing HTTP surfaces, so the
  inconsistency is visible to users rather than being an internal-only detail.

## Scope

- Converge scheduler admin JSON responses onto the canonical `contract` write
  path where the surface is JSON.
- Decide and document the correct 404/405 policy for scheduler admin routes:
  either structured `contract` errors for the full JSON surface or a tightly
  justified exception for true mux-level misses.
- Converge distributed pubsub cluster JSON responses onto the same explicit
  helper path used by the error cases.
- Keep Prometheus success responses as `text/plain`, but remove ad hoc plain-text
  method-mismatch handling if a structured or at least package-consistent path
  is more appropriate.
- Add focused tests that assert status code, content type, and response shape
  for representative success and failure paths.

## Non-Goals

- Do not redesign scheduler job semantics, DLQ behavior, or pubsub cluster
  replication logic.
- Do not change Prometheus exposition format on successful metric responses.
- Do not widen `contract` to absorb cluster- or scheduler-specific business
  policy.

## Expected Files

- `x/scheduler/admin_http.go`
- `x/scheduler/*_test.go`
- `x/pubsub/distributed.go`
- `x/pubsub/*distributed*_test.go`
- `x/pubsub/prometheus.go`
- `x/pubsub/*prometheus*_test.go`

## Validation

```bash
rg -n 'writeJSON\(|json\.NewEncoder\(w\)\.Encode|http\.Error\(|http\.NotFound\(' x/scheduler x/pubsub -g '!**/*_test.go'
go test -timeout 20s ./x/scheduler/... ./x/pubsub/...
go vet ./x/scheduler/... ./x/pubsub/...
```

## Docs Sync

- only if endpoint response behavior or documented admin/export error semantics
  change

## Done Definition

- Scheduler admin handlers no longer rely on a private JSON helper that bypasses
  the canonical contract response path.
- Pubsub distributed HTTP handlers no longer hand-roll JSON success responses.
- Prometheus and cluster/admin failure paths use one intentional, documented
  policy instead of a mix of `http.Error`, raw JSON, and contract errors.
- Focused scheduler and pubsub tests cover representative success and failure
  responses.

## Outcome

- Removed the private scheduler `writeJSON` helper and converged scheduler admin success/error responses onto `contract`.
- Replaced scheduler admin `http.NotFound` branches with structured contract errors for route, job, and dead-letter misses.
- Converged distributed pubsub cluster handlers onto `contract.WriteJSON` / `contract.WriteError`.
- Kept Prometheus success output as `text/plain` while making method mismatch handling structured and test-backed.

## Validation Run

```bash
rg -n 'writeJSON\(|json\.NewEncoder\(w\)\.Encode|http\.Error\(|http\.NotFound\(' x/scheduler x/pubsub -g '!**/*_test.go'
go test -timeout 20s ./x/scheduler/... ./x/pubsub/...
go vet ./x/scheduler/... ./x/pubsub/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```
