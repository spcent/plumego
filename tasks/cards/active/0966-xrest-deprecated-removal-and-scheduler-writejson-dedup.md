# Card 0966: x/rest Deprecated Type Removal and Scheduler writeJSON Deduplication

Milestone:
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P2
State: active
Primary Module: x/rest
Owned Files: x/rest/resource.go, x/scheduler/admin_http.go
Depends On:

## Goal

Remove the explicitly-deprecated legacy types from `x/rest` (they have no external callers
and the deprecation notices have been in place for multiple releases) and replace the
private `writeJSON` function in `x/scheduler/admin_http.go` with `contract.WriteJSON` to
eliminate an undiscovered duplicate that skips the shared buffer pool.

## Problem

### 1. Dead deprecated surface in x/rest

`x/rest/resource.go` contains three types marked `// Deprecated: use contract.*` with no
callers outside the package itself:

| Symbol | Deprecated in favor of |
|---|---|
| `type Response struct` | `contract.Response` / `contract.WriteResponse` |
| `type ErrorResponse struct` | `contract.APIError` / `contract.WriteError` |
| `type JSONWriter struct` (+ all methods) | `contract.WriteResponse` / `contract.WriteError` |

`Response` and `ErrorResponse` have **zero** usages anywhere (`rg 'rest\.Response\b'` and
`rg 'rest\.ErrorResponse\b'` return empty). `JSONWriter` is used only inside the package
itself by `BaseResourceController`.

Additionally, `QueryParams` carries two deprecated fields:

```go
SortBy    string // Deprecated: use Sort
SortOrder string // Deprecated: use Sort
```

These fields are populated by the `QueryBuilder.Parse` method from the legacy `sort_by` /
`sort_order` URL parameters and are only referenced within `resource.go`. There are no
external callers.

**Plan for `JSONWriter`**: inline its four methods directly into `BaseResourceController`,
then delete the type. `BaseResourceController` itself is not deprecated and should remain.

**Plan for `SortBy`/`SortOrder`**: remove the fields from `QueryParams`, remove the
corresponding parse block and the skip-list entries (`"sort_by"`, `"sort_order"`) in
`QueryBuilder.Parse`, and drop the `NormalizeQueryParams` sort-allowlist logic that
references `params.SortOrder`.

### 2. Private `writeJSON` in x/scheduler duplicates contract.WriteJSON

`x/scheduler/admin_http.go` defines a private helper:

```go
func writeJSON(w http.ResponseWriter, status int, payload any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(payload)
}
```

`contract.WriteJSON` does the same but also uses the shared `getJSONBuffer()` pool,
reducing allocations. The scheduler's private version exists for no reason other than
never having been connected to the contract package. It is used in 10+ call sites within
admin_http.go.

**Plan**: replace every call to `writeJSON(w, status, payload)` with
`_ = contract.WriteJSON(w, status, payload)` and delete the private function.

Note: `contract.WriteResponse` is not used here because the scheduler admin API returns
raw JSON objects (not the `{"data":..., "request_id":...}` envelope), and the handler
methods do not receive an `*http.Request`.

## Scope

**x/rest/resource.go**:
- Delete `type Response struct` and `type ErrorResponse struct`.
- Delete `type JSONWriter struct` and all its methods (`WriteJSON`, `WriteError`,
  `WriteResponse`, `WriteAPIError`, `WriteNotImplemented`).
- Remove the `jsonWriter *JSONWriter` field from `BaseResourceController`.
- Inline the four helper calls that `BaseResourceController` delegated through `JSONWriter`:
  `WriteJSON` → `contract.WriteResponse(w, nil, status, data, nil)` (or `contract.WriteJSON`),
  `WriteError` → direct `contract.WriteError` with builder,
  `WriteResponse` → `contract.WriteResponse`,
  `WriteAPIError` → `contract.WriteError`.
- Remove `NewJSONWriter` constructor.
- Remove the `NewBaseResourceController` line that constructs `JSONWriter`.
- Remove `SortBy` and `SortOrder` fields from `QueryParams`.
- Remove the `sort_by` / `sort_order` parse block and skip-list entries in `QueryBuilder.Parse`.
- Remove the `params.SortOrder` reference in the legacy sort block.

**x/scheduler/admin_http.go**:
- Replace all calls `writeJSON(w, status, payload)` with `_ = contract.WriteJSON(w, status, payload)`.
- Delete the private `writeJSON` function.
- Add import for `contract` package (already imported for error building; confirm it covers `WriteJSON`).

## Non-Goals

- Do not remove or deprecate `BaseResourceController` itself — it is not deprecated.
- Do not change `ContextResourceController` or `DBResourceController` — the Ctx-based
  controller migration is a separate larger effort.
- Do not change scheduler response shapes or add request-ID injection — that requires
  refactoring handler method signatures and is out of scope.
- Do not touch `x/rest/spec.go`, `x/rest/entrypoints.go`, `x/rest/resource_db.go`,
  or `x/rest/versioning/`.

## Files

- `x/rest/resource.go`
- `x/rest/routes_test.go` (update if it calls `NewBaseResourceController` in ways that
  rely on internal JSONWriter behaviour)
- `x/scheduler/admin_http.go`

## Tests

Verify no residual symbols remain:

```bash
rg 'rest\.Response\b|rest\.ErrorResponse\b|JSONWriter|NewJSONWriter|\.SortBy\b|\.SortOrder\b' \
  --include='*.go' x/rest/ x/scheduler/
```

Run targeted tests:

```bash
go test -timeout 20s ./x/rest/... ./x/scheduler/...
go vet ./x/rest/... ./x/scheduler/...
```

## Docs Sync

None — the deprecated types were never advertised in docs.

## Done Definition

- `type Response`, `type ErrorResponse`, `type JSONWriter` and `NewJSONWriter` are gone
  from `x/rest/resource.go`.
- `BaseResourceController` still compiles, with all helper methods rewritten to use
  `contract.*` directly.
- `QueryParams.SortBy` and `QueryParams.SortOrder` fields are removed; `QueryBuilder.Parse`
  no longer populates them.
- The private `writeJSON` function in `x/scheduler/admin_http.go` is gone; all its call
  sites use `contract.WriteJSON`.
- All targeted tests and vet pass.

## Outcome

