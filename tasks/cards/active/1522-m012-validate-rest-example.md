# Card 1522

Milestone: M-012
Recipe: specs/change-recipes/update-docs.yaml
Priority: P1
State: active
Primary Module: reference/with-rest
Owned Files:
- `reference/with-rest/internal/handler/create_item.go`
- `reference/with-rest/internal/handler/create_item_test.go`

Goal:
- Add a concrete x/validate usage example to reference/with-rest showing
  Bind[T] with the playground adapter wiring a validated create-item handler.

Scope:
- Add or update reference/with-rest/internal/handler/create_item.go to use
  x/validate.Bind[CreateItemRequest] with a playground.NewValidator() instance
  injected via constructor.
- The handler returns contract.WriteError with the ValidationError on failure
  and contract.WriteResponse on success.
- Write or update create_item_test.go covering: valid request passes, missing
  required field returns 400 with structured error, malformed JSON returns 400.
- Ensure reference/with-rest/go.mod references x/validate/playground as a
  replace directive or direct require (acceptable since reference apps are not
  the main module).

Non-goals:
- Do not change x/validate or x/validate/playground source in this card.
- Do not wire validation globally as middleware.
- Do not change other handlers in reference/with-rest.

Files:
- `reference/with-rest/internal/handler/create_item.go`
- `reference/with-rest/internal/handler/create_item_test.go`

Tests:
- `go build ./reference/with-rest/...`
- `go test -timeout 30s ./reference/with-rest/...`
- `go vet ./reference/with-rest/...`

Docs Sync:
- Update reference/with-rest/README.md to mention validation in the feature list.

Done Definition:
- create_item.go uses Bind[T] and returns contract.WriteError on validation
  failure.
- Three test cases (valid, missing field, bad JSON) pass.
- reference/with-rest builds without errors.

Outcome:
-
