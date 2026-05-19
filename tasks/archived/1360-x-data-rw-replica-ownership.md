---

State: done
id: 1360
title: x/data rw replica ownership
status: done
priority: P2
primary_module: x/data/rw
---

## Goal

Prevent callers from mutating `rw.Cluster` replica state through input or output
slices after cluster construction.

## Scope

- Copy `Config.Replicas` in `New`.
- Copy `ReplicaWeights` before balancer construction when relevant.
- Return a copy from `Replicas()`.
- Add tests proving input and returned slices cannot mutate cluster routing.

## Non-goals

- Hide individual `*sql.DB` handles.
- Change `Close` ownership semantics.
- Add new cluster lifecycle APIs.

## Files

- `x/data/rw/cluster.go`
- `x/data/rw/cluster_test.go`

## Tests

- `go test -timeout 20s ./x/data/rw`
- `go test -race -timeout 60s ./x/data/rw`
- `go vet ./x/data/rw`

## Docs Sync

No docs change expected unless ownership wording is updated.

## Done Definition

- Mutating caller slices after `New` does not affect cluster internals.
- Mutating the slice returned by `Replicas()` does not affect cluster internals.

## Validation

- `go test -timeout 20s ./x/data/rw`
- `go test -race -timeout 60s ./x/data/rw`
- `go vet ./x/data/rw`
