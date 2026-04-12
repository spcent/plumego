# Card 0962: Store File Boundary Helper Ownership Doc Resync

Priority: P3
State: done
Primary Module: store

## Goal

Align the remaining store/file boundary docs so adjacent control-plane guidance
no longer teaches that stable `store/file` owns helper families that were
already pushed out to `x/data/file` and `x/fileapi`.

## Problem

- `docs/modules/x-fileapi/README.md` still says to keep "pure interfaces and
  helpers" in `store/file`.
- The same file still describes `store/file` as owning transport-agnostic
  interfaces, errors, shared types, and "pure helpers".
- `store/module.yaml` still says stable store is focused on "base storage
  contracts and helpers", which lags the tighter stable-root wording used in
  `store/file/file.go`, `store/file/README.md`, and `docs/modules/store/README.md`.

The implementation-side cleanup is done, but the adjacent docs still advertise
the pre-cleanup ownership model.

## Scope

- Resync `docs/modules/x-fileapi/README.md` with the current `store/file`
  boundary: contracts, shared types, and errors stay stable; helper policy does
  not.
- Tighten `store/module.yaml` wording so the stable store manifest no longer
  implies helper-family ownership as part of the canonical root surface.
- Keep the language consistent with the already-updated `store/file` package and
  module docs.

## Non-Goals

- Do not change `store/file` runtime APIs or add/remove symbols.
- Do not widen the card into a broader `x/fileapi` transport refactor.
- Do not reopen tenant/file backend ownership decisions already settled in
  `x/data/file`.

## Files

- `docs/modules/x-fileapi/README.md`
- `store/module.yaml`
- any directly touched store-boundary docs needed for wording consistency

## Tests

- `rg -n "pure interfaces and helpers|pure helpers|base storage contracts and helpers|stable interfaces or pure utilities" docs/modules/x-fileapi/README.md store/module.yaml docs/modules/store/README.md store/file/README.md store/file/file.go`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`

## Docs Sync

- No repo-wide docs sync expected beyond the touched boundary docs and manifest.

## Done Definition

- `docs/modules/x-fileapi/README.md` no longer teaches helper-family ownership
  in `store/file`.
- `store/module.yaml` no longer describes stable store as owning helper
  families.
- Store boundary wording is internally consistent across touched docs.

## Outcome

- Reworded `docs/modules/x-fileapi/README.md` so adjacent transport docs now
  describe `store/file` as owning stable contracts, shared types, and errors
  instead of generic helper families.
- Tightened `store/module.yaml` agent hints to match the already-updated
  `store/file` boundary language.
- Confirmed the touched store/file boundary docs no longer advertise the stale
  helper-ownership model.

## Validation Run

```bash
rg -n "pure interfaces and helpers|pure helpers|base storage contracts and helpers|stable interfaces or pure utilities" docs/modules/x-fileapi/README.md store/module.yaml docs/modules/store/README.md store/file/README.md store/file/file.go
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
```
