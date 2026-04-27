# Card 0341: X Rest Legacy Sort Query Pruning

Priority: P2
State: done
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: x/rest
Depends On: —

## Goal

Finish the remaining `x/rest` query-surface cleanup by removing the old
`sort_by` / `sort_order` compatibility path and the deprecated `QueryParams`
fields that only exist to support it.

## Problem

- `x/rest/resource.go` still keeps deprecated `QueryParams.SortBy` and
  `QueryParams.SortOrder` fields even though the canonical sort model is now the
  `Sort []SortField` slice.
- The query parser still accepts legacy `sort_by` / `sort_order` parameters and
  then rehydrates them back into the canonical `Sort` slice, which means
  `x/rest` still carries a hidden second query vocabulary.
- Repo-wide grep shows no docs, tests, or in-repo callers relying on the legacy
  fields or query parameters, so the compatibility layer now appears to be dead
  tail rather than an actively validated contract.

## Scope

- Remove deprecated `SortBy` and `SortOrder` from `QueryParams`.
- Remove `sort_by` / `sort_order` parsing and skip-list handling from the query
  builder.
- Add or update focused `x/rest` tests so canonical `sort=` parsing remains
  covered after the legacy path disappears.
- Update docs or package comments only if they still mention the old query
  parameters.

## Non-Goals

- Do not redesign pagination, filters, search, or resource route registration.
- Do not change the canonical `sort=field,-other` query format.
- Do not widen this card into broader `x/rest` query DSL redesign.

## Files

- `x/rest/resource.go`
- `x/rest/*_test.go`
- `docs/modules/x-rest/README.md` only if it still mentions the removed query
  params

## Tests

```bash
rg -n 'SortBy|SortOrder|sort_by|sort_order' x/rest docs/modules/x-rest -g '*.*'
go test -timeout 20s ./x/rest/...
go vet ./x/rest/...
```

## Docs Sync

- `docs/modules/x-rest/README.md` or package comments only if they still teach
  `sort_by` / `sort_order`

## Done Definition

- `QueryParams` no longer exports deprecated `SortBy` and `SortOrder` fields.
- `x/rest` no longer parses `sort_by` / `sort_order`.
- Focused `x/rest` tests cover canonical `sort=` parsing without relying on the
  removed compatibility path.
- Repo grep for removed field names and legacy query params is empty outside
  intentional historical archive text.

## Outcome

- Removed deprecated `QueryParams.SortBy` / `SortOrder` from `x/rest`.
- Removed `sort_by` / `sort_order` parsing and the associated special-case skip
  handling from `QueryBuilder.Parse`.
- Added a focused regression test that confirms legacy sort keys no longer feed
  the canonical sort parser.

## Validation Run

```bash
rg -n 'SortBy|SortOrder|sort_by|sort_order' x/rest docs/modules/x-rest -g '*.*'
go test -timeout 20s ./x/rest/...
go vet ./x/rest/...
```
