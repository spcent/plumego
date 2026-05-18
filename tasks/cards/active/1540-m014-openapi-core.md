# Card 1540

Milestone: M-014
Recipe: specs/change-recipes/add-package.yaml
Priority: P1
State: active
Primary Module: x/openapi
Owned Files:
- `x/openapi/openapi.go`
- `x/openapi/openapi_test.go`
- `x/openapi/module.yaml`
- `x/openapi/go.mod`

Goal:
- Create x/openapi with a Generator type that converts router.RouteInfo slices
  and optional Op hint maps into an OpenAPI 3.1 document struct.

Scope:
- Create x/openapi/go.mod with module github.com/spcent/plumego/x/openapi;
  depends on router (for RouteInfo) and contract (for error codes).
- Create x/openapi/openapi.go defining:
  - Op struct: Summary, Description string; Tags []string; Params []Param;
    Body, Responses map entries.
  - Param struct: Name, In, Description string; Required bool; Schema Schema.
  - Schema struct: Type, Format string; Properties map[string]Schema; Ref string.
  - Document struct: OpenAPI string; Info Info; Paths map[string]PathItem.
  - Generator struct with New() constructor.
  - Generate(routes []router.RouteInfo, hints map[string]Op) Document — merges
    route metadata with hints; routes without hints get a default 200 response.
  - PathParam, QueryParam, HeaderParam constructor helpers for Param.
  - String, Integer, Boolean, Array Schema constants.
- Create x/openapi/module.yaml with status = experimental, owner = api,
  allowed_imports listing only router, contract, and stdlib.
- Write x/openapi/openapi_test.go covering:
  - empty route list returns document with empty Paths.
  - static route without hint appears with default 200 response.
  - parameterised route (:id) generates correct path template ({id}).
  - route with Op hint uses hint Summary and Params.
  - multiple routes produce correct Paths map keys.

Non-goals:
- Do not add serialisation to this card (that is card 1541).
- Do not add the CLI command in this card (that is card 1542).
- Do not add external OpenAPI library dependencies.
- Do not add to the main module go.mod.

Files:
- `x/openapi/openapi.go`
- `x/openapi/openapi_test.go`
- `x/openapi/module.yaml`
- `x/openapi/go.mod`

Tests:
- `go test -race -timeout 60s ./x/openapi/...`
- `go vet ./x/openapi/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- none at this card; reference example added in card 1542.

Done Definition:
- x/openapi/openapi.go compiles; Generator.Generate returns a valid Document.
- All five test cases pass with `go test -race`.
- x/openapi imports only router, contract, and stdlib.
- `go run ./internal/checks/dependency-rules` exits 0.

Outcome:
-
