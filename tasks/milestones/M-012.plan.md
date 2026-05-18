# Plan for M-012: Input Validation Bridge

Milestone: `M-012`
Objective: Ship x/validate with a Validator interface and generic Bind[T] helper
that eliminates common binding-and-validation boilerplate, plus an optional
x/validate/playground sub-package that adapts go-playground/validator v10 without
adding that dependency to the main module.
Constraints: x/validate depends only on stdlib and contract stable root, playground
sub-package has its own go.mod, go-playground/validator stays out of main module
go.mod, validation errors use contract.APIError via contract.WriteError only.
Affected Modules: x/validate, contract (read-only), reference/with-rest.

## Phase Map

- Phase 1: Orient — confirm x/validate/ does not exist; read contract/module.yaml
  and contract/errors.go to understand APIError shape.
- Phase 2: Implement (parallel) — write core validate package, playground adapter,
  and reference example concurrently.
- Phase 3: Test — write validate_test.go and adapter_test.go covering all
  positive and negative paths.
- Phase 4: Validate and Ship — run acceptance criteria, commit.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1520 | Create x/validate core package with Validator interface and Bind[T] | x/validate | `x/validate/validate.go`, `x/validate/module.yaml` | M-008 | `go test ./x/validate/...`, `go vet ./x/validate/...` |
| 1521 | Create x/validate/playground adapter wrapping go-playground/validator v10 | x/validate | `x/validate/playground/adapter.go`, `x/validate/playground/go.mod` | 1520 | `go test ./x/validate/playground/...` |
| 1522 | Add x/validate usage example to reference/with-rest | reference/with-rest | `reference/with-rest/internal/handler/validated_handler.go` | 1520 | `go build ./reference/with-rest/...` |

## Dependency Edges

- `1520 -> 1521`
- `1520 -> 1522`

## Parallel Groups

- Group A: card 1520 — must complete first as the interface definition.
- Group B (parallel after A): cards 1521 and 1522 — independent consumers of the interface.
- Group C (sequential after B): tests for both 1521 and 1522.

## Risk Register

- Risk: Go generics constraint on Bind[T] causes vet failures on older toolchain.
  Mitigation: require go 1.21+ in x/validate/go.mod; CLAUDE.md already sets go 1.24.0.
- Risk: contract.APIError shape changes between now and card execution.
  Mitigation: card 1520 reads contract/errors.go before writing any code; stop and
  flag if the shape differs from the spec.

## Verification Strategy

- Card-level checks: `go test ./x/validate/...` after 1520; `go test ./x/validate/playground/...`
  after 1521; `go build ./reference/with-rest/...` after 1522.
- Negative-path coverage: validate_test.go must exercise missing field, type mismatch,
  empty body, and oversized body; acceptance criteria will fail if coverage is absent.
- Dependency audit: run `go run ./internal/checks/dependency-rules` to confirm no
  stable root imports x/validate and x/validate does not import x/* other than contract.

## Exit Condition

- all three cards completed and tests written
- x/validate/validate.go with Validator interface and Bind[T] exists
- x/validate/playground/go.mod is a separate module
- reference/with-rest shows a validated handler
- verify report shows pass
- milestone acceptance criteria ready for PR packaging
