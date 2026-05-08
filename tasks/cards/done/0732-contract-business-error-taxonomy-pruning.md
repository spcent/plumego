# Card 0732

Milestone: v1
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/error_codes.go
- contract/errors_test.go
- x/rest/resource_db.go
- x/scheduler/admin_http.go
- docs/modules/contract/README.md
- contract/module.yaml
Depends On:
- 0731

Goal:
Remove domain/business error taxonomy from stable `contract` so the package owns transport primitives rather than product semantics.

Scope:
- Enumerate all Go callers of `CategoryBusiness`, `TypeInvalidState`, `TypeInsufficientFunds`, `TypeOperationNotAllowed`, `CodeInvalidState`, `CodeInsufficientFunds`, and `CodeOperationNotAllowed` before editing.
- Remove those exported business-oriented symbols from `contract`.
- Migrate `x/*` callers to explicit transport categories, statuses, and codes that already belong to their owning extension surface.
- Update tests and contract public-surface docs/manifests.

Non-goals:
- Do not introduce a new extension-wide business error framework.
- Do not change HTTP response envelopes.
- Do not refactor unrelated `x/rest` or `x/scheduler` behavior.

Files:
- contract/errors.go
- contract/error_codes.go
- contract/errors_test.go
- x/rest/resource_db.go
- x/scheduler/admin_http.go
- docs/modules/contract/README.md
- contract/module.yaml

Tests:
- rg -n --glob '*.go' 'CategoryBusiness|TypeInvalidState|TypeInsufficientFunds|TypeOperationNotAllowed|CodeInvalidState|CodeInsufficientFunds|CodeOperationNotAllowed' .
- go test -timeout 20s ./contract/... ./x/rest/... ./x/scheduler/...
- go build ./...

Docs Sync:
- Sync `contract/module.yaml` and `docs/modules/contract/README.md` with the reduced public surface.

Done Definition:
- The removed business taxonomy symbols no longer appear in Go code.
- Extension callers still produce structured errors with explicit status, category, code, and message.
- Targeted tests and build validation pass.

Outcome:
- Enumerated all old business taxonomy callers before editing.
- Removed `CategoryBusiness`, business-oriented `ErrorType` constants, and business-oriented `Code*` constants from `contract`.
- Migrated `x/rest` hook rejection paths to explicit transport status/category/code values.
- Migrated scheduler trigger failure to explicit 422 client error metadata.
- Updated contract tests that asserted the removed category mapping.

Validation:
- rg -n --glob '*.go' 'CategoryBusiness|TypeInvalidState|TypeInsufficientFunds|TypeOperationNotAllowed|CodeInvalidState|CodeInsufficientFunds|CodeOperationNotAllowed' .
- go test -timeout 20s ./contract/... ./x/rest/... ./x/scheduler/...
- go build ./...
- go run ./internal/checks/module-manifests
