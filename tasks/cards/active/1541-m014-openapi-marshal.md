# Card 1541

Milestone: M-014
Recipe: specs/change-recipes/add-package.yaml
Priority: P1
State: active
Primary Module: x/openapi
Owned Files:
- `x/openapi/marshal.go`
- `x/openapi/marshal_test.go`

Goal:
- Add JSON and YAML serialisation of the OpenAPI Document struct to x/openapi,
  enabling the CLI command and user code to write spec files in either format.

Scope:
- Create x/openapi/marshal.go defining:
  - MarshalJSON(doc Document) ([]byte, error) — uses encoding/json with
    MarshalIndent for readability.
  - MarshalYAML(doc Document) ([]byte, error) — uses a minimal YAML encoder
    implemented over encoding/json (marshal to JSON map, convert to YAML)
    with no external YAML library dependency.
  - WriteJSON(w io.Writer, doc Document) error
  - WriteYAML(w io.Writer, doc Document) error
- Write x/openapi/marshal_test.go covering:
  - MarshalJSON produces valid JSON with correct Content-Type implication.
  - MarshalYAML produces valid YAML (key: value format, no JSON-style braces).
  - Round-trip: unmarshal the JSON output back to Document and compare key fields.
  - Empty Document serialises without error.

Non-goals:
- Do not add gopkg.in/yaml.v3 or any external YAML library to x/openapi/go.mod.
- Do not add XML serialisation.
- Do not validate the OpenAPI spec for correctness (that is a user responsibility).

Files:
- `x/openapi/marshal.go`
- `x/openapi/marshal_test.go`

Tests:
- `go test -race -timeout 60s ./x/openapi/...`
- `go vet ./x/openapi/...`

Docs Sync:
- none; format flag documented in CLI card 1542.

Done Definition:
- MarshalJSON and MarshalYAML both produce parseable output.
- Round-trip test passes.
- No external dependencies added to x/openapi/go.mod beyond those in card 1540.

Outcome:
-
