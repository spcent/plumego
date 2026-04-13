# Card 0979: X/Fileapi Handler Wrapper Removal and Codegen Error Code Canonicalization

Priority: P3
State: active
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/fileapi, cmd/plumego/internal/codegen
Depends On: 0972

## Goal

Remove the redundant local helper wrappers in `x/fileapi/handler.go` that
obscure the canonical contract usage, fix the non-canonical error code
generation inside those helpers, and update the `cmd/plumego/internal/codegen`
templates to generate canonical error code constants instead of lowercase
freeform strings.

## Problem

### x/fileapi/handler.go — local wrapper indirection

`x/fileapi/handler.go` (lines 257–276) defines three private helper methods on
`*Handler`:

```go
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
    _ = contract.WriteJSON(w, status, data)   // thin passthrough — no value
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
    contract.WriteError(w, r, contract.NewErrorBuilder().  // missing _ = discard
        Status(status).
        Code(http.StatusText(status)).          // "Bad Request", "Not Found" — not canonical
        Message(message).
        Category(contract.CategoryForStatus(status)).
        Build())
}

func (h *Handler) writeMetadataError(w http.ResponseWriter, r *http.Request, err error) {
    if err == storefile.ErrNotFound { ... }   // uses writeError indirectly
}
```

Issues:
1. `writeJSON` adds nothing over calling `contract.WriteJSON` directly;
   16 call sites reference it unnecessarily.
2. `writeError` uses `http.StatusText(status)` for the error code, producing
   human-readable codes like `"Bad Request"` and `"Not Found"` instead of the
   canonical `CodeBadRequest = "BAD_REQUEST"` and
   `CodeResourceNotFound = "RESOURCE_NOT_FOUND"`.
3. The `contract.WriteError` call inside `writeError` is missing `_ =`.

### cmd/plumego/internal/codegen/codegen.go — template error codes

The `generateHandlerMethodCode` function (lines 300–402) produces handler
templates with hardcoded lowercase codes:

- `Code("%s_not_found")` → should use `contract.CodeResourceNotFound`
- `Code("invalid_json")` → should use `contract.CodeInvalidJSON`
- `Code("create_%s_failed")`, `Code("update_%s_failed")`,
  `Code("delete_%s_failed")` → should use `contract.CodeInternalError`

Because these are code generation templates, the non-canonical codes propagate
into every project that runs `plumego generate`. The `"invalid_json"` case is
the most critical since `contract.CodeInvalidJSON = "INVALID_JSON"` already
exists as an exact match.

## Scope

### x/fileapi

- Replace all `h.writeJSON(...)` call sites with `_ = contract.WriteJSON(...)`.
- Replace all `h.writeError(...)` call sites with direct
  `_ = contract.WriteError(w, r, contract.NewErrorBuilder()...).Build())` using
  the correct `contract.Code*` constant for each status (e.g.
  `CodeBadRequest`, `CodeResourceNotFound`, `CodeInternalError`).
- Update `writeMetadataError` inline logic similarly.
- Delete the `writeJSON`, `writeError`, and `writeMetadataError` helpers once
  all callers are migrated.

### cmd/plumego/internal/codegen

- In `generateHandlerMethodCode`, replace:
  - `Code("%s_not_found")` → `Code(contract.CodeResourceNotFound)` (with `%s`
    in the Message only)
  - `Code("invalid_json")` → `Code(contract.CodeInvalidJSON)`
  - `Code("create_%s_failed")`, `Code("update_%s_failed")`,
    `Code("delete_%s_failed")` → `Code(contract.CodeInternalError)`
- Add `"github.com/spcent/plumego/contract"` to the generated imports if not
  already present in the handler template.
- Update codegen tests that assert on the old generated code strings.

## Non-Goals

- Do not change file upload, download, or metadata storage logic in x/fileapi.
- Do not change codegen CLI flags, command names, or overall template structure.
- Do not change the DELETE handler's `w.WriteHeader(http.StatusNoContent)` —
  204 No Content by definition has no body, so the raw WriteHeader is acceptable
  there.

## Files

- `x/fileapi/handler.go`
- `x/fileapi/*_test.go`
- `cmd/plumego/internal/codegen/codegen.go`
- `cmd/plumego/internal/codegen/*_test.go`

## Tests

```bash
rg -n 'writeJSON\|writeError\|writeMetadataError\|http\.StatusText' x/fileapi -g '*.go'
rg -n '"invalid_json"\|"%s_not_found"\|"create_%s_failed"\|"update_%s_failed"\|"delete_%s_failed"' cmd/plumego/internal/codegen -g '*.go'
go test -timeout 20s ./x/fileapi/... ./cmd/plumego/internal/codegen/...
go vet ./x/fileapi/... ./cmd/plumego/internal/codegen/...
```

## Docs Sync

None required.

## Done Definition

- `x/fileapi/handler.go` contains no `writeJSON`, `writeError`, or
  `writeMetadataError` helpers; all call sites use contract functions directly.
- No `http.StatusText(...)` call appears in `x/fileapi/handler.go`.
- Generated handler code uses `contract.CodeResourceNotFound`,
  `contract.CodeInvalidJSON`, and `contract.CodeInternalError` instead of
  freeform lowercase strings.
- All targeted package tests pass; `go vet` is clean.

## Outcome

