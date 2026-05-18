# Card 1520

Milestone: M-012
Recipe: specs/change-recipes/add-package.yaml
Priority: P1
State: active
Primary Module: x/validate
Owned Files:
- `x/validate/validate.go`
- `x/validate/validate_test.go`
- `x/validate/module.yaml`
- `x/validate/go.mod`

Goal:
- Create x/validate with a Validator interface, generic Bind[T] and BindJSON[T]
  helpers, and a module.yaml declaring the new package as experimental.

Scope:
- Create x/validate/validate.go defining:
  - Validator interface: Validate(v any) error
  - Bind[T any](r *http.Request, v Validator) (T, error) — decodes JSON body
    then calls v.Validate
  - BindJSON[T any](r *http.Request) (T, error) — decodes without validation
  - ValidationError type wrapping Validator errors as contract.APIError with
    TypeBadRequest and Code "VALIDATION_ERROR"
- Create x/validate/module.yaml with status = experimental, owner = api,
  forbidden_imports listing x/* packages other than contract.
- Create x/validate/go.mod with module github.com/spcent/plumego/x/validate.
- Write x/validate/validate_test.go covering:
  - valid JSON input with passing validation
  - missing required field returns ValidationError
  - malformed JSON returns error before validation
  - empty body returns error
  - Validator returning nil passes through

Non-goals:
- Do not implement a validation tag language or struct tag parser.
- Do not add go-playground/validator as a dependency (that is card 1521).
- Do not add middleware; validation is explicit at call sites.
- Do not add x/validate to the main module go.mod.

Files:
- `x/validate/validate.go`
- `x/validate/validate_test.go`
- `x/validate/module.yaml`
- `x/validate/go.mod`

Tests:
- `go test -race -timeout 60s ./x/validate/...`
- `go vet ./x/validate/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- none (module.yaml is the primary declaration; primer added in a later card)

Done Definition:
- x/validate/validate.go compiles with Go 1.21+ generics.
- Bind[T] correctly propagates ValidationError for failing validation.
- All five test cases in validate_test.go pass with `go test -race`.
- x/validate/go.mod is a separate module not listed in main go.mod.
- `go run ./internal/checks/dependency-rules` exits 0.

Outcome:
-
