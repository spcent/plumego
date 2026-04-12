# Card 0961: Extension Context Accessor Contract Normalization

Priority: P1
State: done
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: multi-module extension context helpers
Depends On: —

## Goal

Bring extension-package context helpers back into the repository's canonical
`With{Type}` + `{Type}FromContext` contract with unexported zero-value struct
keys, so extension code stops drifting away from the accessor rules already
documented in `AGENTS.md`.

## Problem

- `x/data/rw/policy.go` uses `ctxKey int` plus `iota` constants for
  `ForcePrimary`, `PreferReplica`, and `MarkInTransaction`, which violates the
  repo rule that context keys use unexported zero-value struct types.
- The same file also exports `ForcePrimary` / `IsForcePrimary` style helpers
  instead of the canonical `WithUsePrimary` / `UsePrimaryFromContext`-style
  pairing.
- `x/rest/versioning/version.go` stores API version in context with a private
  struct key, but the public accessor is named `GetVersion` rather than
  `VersionFromContext`, so the package still teaches a second accessor
  convention.
- `x/scheduler/types.go` defines `JobIDFromContext`, `JobAttemptFromContext`,
  `JobScheduledAtFromContext`, and `JobLastErrorFromContext`, but the matching
  `With*` helpers do not exist; context population happens inline inside
  `scheduler.go`, which hides the write side of the contract.

## Scope

- Replace `x/data/rw`'s integer context keys with unexported zero-value struct
  keys.
- Rename or wrap public `x/data/rw` context helpers so the public API follows
  the canonical `With*`/`*FromContext` shape; remove deprecated aliases after
  migrating callers.
- Rename `x/rest/versioning.GetVersion` to the canonical accessor name and
  update all examples/docs in the same change.
- Introduce symmetric `WithJobID`, `WithJobAttempt`, `WithJobScheduledAt`, and
  `WithJobLastError` helpers in `x/scheduler`, and route internal writes
  through them instead of inline `context.WithValue`.
- Follow the exported-symbol completeness protocol for any removed aliases.

## Non-Goals

- Do not redesign version-negotiation strategy or scheduler execution logic.
- Do not change read/write routing semantics in `x/data/rw`.
- Do not move extension context state into stable roots.

## Expected Files

- `x/data/rw/policy.go`
- `x/data/rw/*_test.go`
- `x/rest/versioning/version.go`
- `x/rest/versioning/*_test.go`
- `x/scheduler/types.go`
- `x/scheduler/scheduler.go`
- `x/scheduler/*_test.go`

## Validation

```bash
rg -n 'ctxKey|GetVersion\\(|context\\.WithValue\\(ctx, job|ForcePrimary\\(|PreferReplica\\(|MarkInTransaction\\(' x/data/rw x/rest/versioning x/scheduler . -g '*.go'
go test -timeout 20s ./x/data/rw/... ./x/rest/versioning/... ./x/scheduler/...
go vet ./x/data/rw/... ./x/rest/versioning/... ./x/scheduler/...
```

## Docs Sync

- docs or package comments that still teach `GetVersion` or the old helper
  naming once the canonical accessor pair is introduced

## Done Definition

- Extension context keys use unexported zero-value struct types instead of
  integer key enums.
- Public context helpers in the touched extension packages follow the canonical
  `With*` + `*FromContext` naming contract.
- `x/scheduler` no longer writes job context values inline without symmetric
  helper functions.
- Residual references to removed helper names are empty and focused tests pass.

## Outcome

- Normalized `x/data/rw` onto `WithForcePrimary`, `WithPreferReplica`, `WithInTransaction` and matching `*FromContext` accessors backed by inline zero-value struct keys.
- Replaced `x/rest/versioning.GetVersion` with the canonical `VersionFromContext` / `WithVersion` pair.
- Added symmetric scheduler context attachment helpers and removed inline `context.WithValue` usage for job metadata.
- Updated associated docs, testdata, and module manifests so the control plane now teaches the normalized accessor contract.

## Validation Run

```bash
rg -n 'ctxKey|GetVersion\(|context\.WithValue\(ctx, job|ForcePrimary\(|PreferReplica\(|MarkInTransaction\(' x/data/rw x/rest/versioning x/scheduler . -g '*.go'
go test -timeout 20s ./x/data/rw/... ./x/rest/versioning/... ./x/scheduler/...
go vet ./x/data/rw/... ./x/rest/versioning/... ./x/scheduler/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```
