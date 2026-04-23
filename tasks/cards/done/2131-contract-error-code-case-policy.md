# Card 2131: Contract Error Code Case Policy
Milestone: none
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: medium
State: active
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
- contract/context_extended_test.go
- docs/modules/contract/README.md
Depends On: none

Goal:
Make the contract package's error-code case policy explicit and stop tests from
teaching lowercase custom codes as canonical examples. Repository style expects
structured, uppercase error codes, but `ErrorBuilder.Code` currently preserves
whatever string a caller passes and some contract tests still assert lowercase
examples.

Scope:
- Decide and document the stable-root policy for explicit custom error codes:
  preserve caller input with documented uppercase expectation, or normalize in a
  carefully test-covered way.
- Update contract tests so canonical examples use package constants or uppercase
  code strings.
- If behavior changes, enumerate all affected call sites before editing and keep
  the change inside the contract package unless a caller assertion must be
  updated.
- Preserve the single canonical error write path through `contract.WriteError`.

Non-goals:
- Do not add per-feature error helpers.
- Do not change response envelope structure.
- Do not make stable roots import extension packages.

Tests:
- go test -race -timeout 60s ./contract
- go test -timeout 20s ./contract
- go vet ./contract

Docs Sync:
Update `docs/modules/contract/README.md` with the explicit error-code policy if
the docs do not already state it.

Done Definition:
- Contract tests no longer present lowercase custom error codes as canonical
  examples.
- The package has an explicit, documented behavior for caller-supplied error
  code casing.
- The contract validation commands pass.
- No alternate error-construction path is introduced.

Outcome:
Kept `ErrorBuilder.Code` behavior stable by preserving caller input, then made
the policy explicit in code comments and contract module docs: use `Code*`
constants or uppercase stable strings. Updated contract tests to use canonical
constants instead of lowercase custom examples.

Validation:
- `go test -race -timeout 60s ./contract`
- `go test -timeout 20s ./contract`
- `go vet ./contract`
- `rg -n "invalid_request|bad_request" contract/errors_test.go contract/context_extended_test.go docs/modules/contract/README.md contract/errors.go`
  returned no matches.
