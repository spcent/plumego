# Card 0357: Freeform Error Code Strings in x/ Production Modules

Priority: P2
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/rest, x/websocket, x/frontend
Depends On: 0972, 0980

## Goal

Several production `x/` modules still use raw lowercase string literals in
`.Code("…")` calls instead of the `contract.Code*` constants introduced in
card 0972. Replace every freeform code that has a canonical counterpart.

## Affected Files and Mappings

### `x/websocket/websocket.go` (1 site)

| Line | Freeform | Replace with |
|---|---|---|
| 130 | `"read_body_failed"` | `contract.CodeInternalError` |

### `x/rest/resource_db.go` (20 sites across 5 handler methods)

| Freeform | Occurrences | Replace with |
|---|---|---|
| `"invalid_list_request"` | 1 (Index) | `contract.CodeInvalidRequest` |
| `"list_failed"` | 1 | `contract.CodeInternalError` |
| `"transform_collection_failed"` | 1 | `contract.CodeInternalError` |
| `"after_list_failed"` | 1 | `contract.CodeInternalError` |
| `"missing_id"` | 2 (Show, Update, Delete) | `contract.CodeBadRequest` |
| `"not_found"` | 3 (Show, Update, Delete) | `contract.CodeResourceNotFound` |
| `"show_failed"` | 1 | `contract.CodeInternalError` |
| `"transform_failed"` | 3 | `contract.CodeInternalError` |
| `"invalid_request"` | 2 (Create, Update) | `contract.CodeInvalidRequest` |
| `"validation_failed"` | 2 (Create, Update) | `contract.CodeValidationError` |
| `"before_create_failed"` | 1 | `contract.CodeInternalError` |
| `"create_failed"` | 1 | `contract.CodeInternalError` |
| `"after_create_failed"` | 1 | `contract.CodeInternalError` |
| `"before_update_failed"` | 1 | `contract.CodeInternalError` |
| `"update_failed"` | 1 | `contract.CodeInternalError` |
| `"after_update_failed"` | 1 | `contract.CodeInternalError` |
| `"before_delete_failed"` | 1 | `contract.CodeInternalError` |
| `"delete_failed"` | 1 | `contract.CodeInternalError` |
| `"after_delete_failed"` | 1 | `contract.CodeInternalError` |

Note: `missing_id` count across Index/Show/Update/Delete is 2 (Show and Update
each have one; Delete also has one = 3 total).

### `x/rest/resource.go` (1 site)

| Line | Freeform | Replace with |
|---|---|---|
| 107 | `"not_implemented"` | `contract.CodeMethodNotAllowed` |

### `x/frontend/frontend.go` (2 sites)

| Line | Freeform | Replace with |
|---|---|---|
| 349 | `"method_not_allowed"` | `contract.CodeMethodNotAllowed` |
| 531 | `"internal_error"` | `contract.CodeInternalError` |

### `cmd/plumego/internal/scaffold/scaffold.go` (4 sites)

This is scaffolded sample code; align with canonical constants for consistency:

| Freeform | Replace with |
|---|---|
| `"list_users_failed"` | `contract.CodeInternalError` |
| `"user_not_found"` | `contract.CodeResourceNotFound` |
| `"invalid_json"` | `contract.CodeInvalidJSON` |
| `"create_user_failed"` | `contract.CodeInternalError` |

## Out of Scope

Domain-specific devserver codes in `cmd/plumego/internal/devserver/` such as
`"build_failed"`, `"app_restart_failed"`, `"pprof_fetch_failed"`, etc. These
describe internal tooling operations that have no canonical API-level counterpart
and are intentionally kept as descriptive freeform strings.

Similarly, `x/websocket/server.go` `"JOIN_DENIED"` and `x/scheduler/admin_http.go`
`"TRIGGER_FAILED"` are domain-specific and may remain.

## Tests

```bash
go test -timeout 20s ./x/rest/... ./x/websocket/... ./x/frontend/...
go vet ./x/rest/... ./x/websocket/... ./x/frontend/...
```

## Done Definition

- No lowercase freeform `.Code("…")` string appears in any of the listed files.
- All test suites pass; `go vet` is clean.

## Outcome

Completed. All freeform lowercase code strings in listed files replaced with canonical `contract.Code*` constants:
- `x/websocket/websocket.go`: `"read_body_failed"` → `CodeInternalError`
- `x/rest/resource_db.go`: all 20 sites replaced with canonical constants
- `x/rest/resource.go`: `"not_implemented"` → `CodeMethodNotAllowed` (now uses `TypeNotImplemented` via card 0984)
- `x/frontend/frontend.go`: `"method_not_allowed"` → `TypeMethodNotAllowed`, `"internal_error"` → `CodeInternalError`
- `cmd/plumego/internal/scaffold/scaffold.go`: 4 sites replaced
All tests pass; `go vet` clean.
