# Card 0907

Priority: P2
State: active
Primary Module: contract
Owned Files:
- `contract/errors.go`
Depends On: —

Goal:
- Rename `ErrType*` constants to `Type*` to align with the `Category*` / `Severity*` naming convention used by the other two `ErrorType` peers.

Problem:
The three error classification types in `contract` each have a family of named constants.
The naming pattern is inconsistent:

| Type          | Constants         | Example              |
|---------------|-------------------|----------------------|
| `ErrorCategory` | `Category*`       | `CategoryClient`     |
| `ErrorSeverity` | `Severity*`       | `SeverityInfo`       |
| `ErrorType`     | `ErrType*`        | `ErrTypeValidation`  | ← inconsistent

The `ErrType*` prefix mixes an error-sentinel convention (`Err` prefix) with what is
actually a classification constant (not a sentinel error value). Every other constant in
the package that classifies rather than represents an error uses the type-name without
the `Err` prefix: `Category*`, `Severity*`. Following the same pattern, `ErrorType`
constants should use the `Type*` prefix.

Fix:
- Rename all `ErrType*` constants to `Type*`:
  - `ErrTypeValidation`    → `TypeValidation`
  - `ErrTypeRequired`      → `TypeRequired`
  - `ErrTypeInvalidFormat` → `TypeInvalidFormat`
  - `ErrTypeOutOfRange`    → `TypeOutOfRange`
  - `ErrTypeDuplicate`     → `TypeDuplicate`
  - `ErrTypeUnauthorized`  → `TypeUnauthorized`
  - `ErrTypeForbidden`     → `TypeForbidden`
  - `ErrTypeInvalidToken`  → `TypeInvalidToken`
  - `ErrTypeExpiredToken`  → `TypeExpiredToken`
  - `ErrTypeNotFound`      → `TypeNotFound`
  - `ErrTypeConflict`      → `TypeConflict`
  - `ErrTypeAlreadyExists` → `TypeAlreadyExists`
  - `ErrTypeGone`          → `TypeGone`
  - `ErrTypeInternal`      → `TypeInternal`
  - `ErrTypeUnavailable`   → `TypeUnavailable`
  - `ErrTypeTimeout`       → `TypeTimeout`
  - `ErrTypeRateLimited`   → `TypeRateLimited`
  - `ErrTypeMaintenance`   → `TypeMaintenance`
  - `ErrTypeInvalidState`        → `TypeInvalidState`
  - `ErrTypeInsufficientFunds`   → `TypeInsufficientFunds`
  - `ErrTypeOperationNotAllowed` → `TypeOperationNotAllowed`
- Update the `errorTypeLookup` map keys.
- Grep all callers: `grep -rn 'ErrType' . --include='*.go'`
- Update all call sites.

Non-goals:
- Do not rename `ErrorType` (the type itself).
- Do not change wire values (the string constants like `"validation_error"` stay the same).
- Do not rename `Category*` or `Severity*` constants.

Files:
- `contract/errors.go`
- All caller files found by grep (expect 140+ sites across middleware, x/*)

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/... ./middleware/... ./x/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- No `ErrType*` identifier exists anywhere in the codebase.
- `grep -rn 'ErrType' . --include='*.go'` (excluding tasks) returns empty.
- `go build ./...` passes.
- All tests pass.

Outcome:
State: done
- All 21 `ErrType*` constants renamed to `Type*` across `contract/errors.go`.
- `errorTypeLookup` map keys updated.
- ~140 call sites across `middleware/`, `x/`, and `contract/` updated via bulk sed.
- `grep -rn 'ErrType' . --include='*.go'` returns empty (outside task cards).
- `go test -timeout 20s ./...` passes.
