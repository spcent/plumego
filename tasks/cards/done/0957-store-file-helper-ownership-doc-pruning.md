# Card 0957: Store File Helper Ownership Doc Pruning

Priority: P2
State: done
Primary Module: store

## Goal

Resync the stable `store/file` docs and package comments so they describe the
current exported surface accurately and stop advertising helper ownership that
 no longer exists in the stable root.

## Problem

- `store/file/file.go` still says the package provides file storage
  "contracts and helpers".
- `docs/modules/store/README.md` still says `store/file` owns "path/file
  helpers".
- `store/file/README.md` still describes path/id helpers as extension-owned,
  which only makes sense if stable `store/file` were still the helper owner of
  record.
- The live stable surface under `store/file` is now interfaces, shared types,
  and error contracts; there is no exported path/id helper family left in the
  package.

That leaves readers with a misleading picture of stable file ownership:
`store/file` looks broader in the docs than it is in code.

## Scope

- Remove stale "helpers" or "path/file helpers" ownership claims from
  `store/file` package/docs when they no longer match exported code.
- Reword the stable boundary around `store/file` to emphasize contracts, shared
  types, and errors only.
- Keep extension ownership wording clear for backend-specific behavior,
  metadata, transport, and any helper policy that lives outside the stable root.

## Non-Goals

- Do not add new helper APIs back into `store/file`.
- Do not move implementation code between `store/file`, `x/data/file`, and
  `x/fileapi`.
- Do not change runtime behavior in this card unless a doc mismatch reveals a
  real stable-surface bug.

## Files

- `store/file/file.go`
- `store/file/README.md`
- `docs/modules/store/README.md`

## Tests

- `go test -timeout 20s ./store/file`
- `go test -race -timeout 60s ./store/file`
- `go vet ./store/file`

## Docs Sync

- No repo-wide doc sync expected beyond the touched store docs unless the
  wording mismatch appears elsewhere during implementation.

## Done Definition

- `store/file` package comments and docs no longer claim stable helper
  ownership that is absent from the code.
- Stable store docs describe `store/file` as a contract/type/error layer.
- Extension ownership wording for file helper policy is explicit and
  non-contradictory.

## Outcome

- Reworded the `store/file` package comment to describe contracts, shared types,
  and errors instead of non-existent stable helpers.
- Tightened `store/file/README.md` so helper policy is clearly extension-owned.
- Updated `docs/modules/store/README.md` to remove the stale
  "`store/file` owns path/file helpers" claim.

## Validation Run

```bash
go test -timeout 20s ./store/file
go test -race -timeout 60s ./store/file
go vet ./store/file
```
