# Card 0744

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
- contract/module.yaml
- docs/modules/contract/README.md
Depends On:
- 0743

Goal:
Freeze the error taxonomy matrix across category, status, type, and code so extension callers have one stable interpretation.

Scope:
- Add regression coverage for all public error types and category/status mapping.
- Document timeout, validation, auth, rate-limit, client, and server semantics.
- Define extension code rules without adding feature-specific codes to `contract`.

Non-goals:
- Do not add business error categories.
- Do not add extension-specific error constants.
- Do not change the response envelope.

Files:
- contract/errors.go
- contract/errors_test.go
- contract/module.yaml
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update contract docs and manifest guardrails for taxonomy rules.

Done Definition:
- Every public error type has a canonical status/category/code assertion.
- Extension code ownership guidance is explicit.
- Contract tests and vet pass.

Outcome:
- Added table-driven coverage for every public `ErrorType` canonical status, category, and code.
- Added a fail-closed assertion for unknown `ErrorType.Meta()`.
- Documented taxonomy categories, representative statuses, and extension-owned custom code rules.
- Added manifest guardrails for extension code taxonomy behavior.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
