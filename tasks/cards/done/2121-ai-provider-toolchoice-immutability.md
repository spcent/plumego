# Card 2121: AI Provider ToolChoice Immutability

Milestone: none
Recipe: specs/change-recipes/symbol-change.yaml
Priority: medium
State: done
Primary Module: x/ai
Owned Files:
- x/ai/provider/provider.go
- x/ai/provider/provider_test.go
- docs/modules/x-ai/README.md
Depends On: none

Goal:
- Remove internal reliance on exported mutable `ToolChoice` package variables in the stable `x/ai/provider` API surface.
- Provide a canonical immutable construction path for common tool-choice values.

Scope:
- Follow the symbol-change completeness protocol before changing `ToolChoiceAuto`, `ToolChoiceNone`, or `ToolChoiceAny`.
- Introduce canonical constructor functions or immutable value helpers for auto, none, and any tool choices.
- Update internal tests and examples to use the canonical immutable path.
- Document remaining compatibility debt if public mutable variables cannot be removed without explicit stable API approval.

Non-goals:
- Do not change provider request/response semantics.
- Do not change streaming, session, tool registry, or adapter behavior.
- Do not introduce hidden provider globals or implicit provider registration.

Files:
- `x/ai/provider/provider.go`: add immutable tool-choice helpers and reduce mutable global usage.
- `x/ai/provider/provider_test.go`: cover helper values and ensure mutations cannot affect canonical defaults.
- `docs/modules/x-ai/README.md`: document the canonical tool-choice path if public guidance changes.

Tests:
- `go test -race -timeout 60s ./x/ai/provider`
- `go test -timeout 20s ./x/ai/provider`
- `go vet ./x/ai/provider`

Docs Sync:
- Required if public provider guidance changes or compatibility debt is documented.

Done Definition:
- Internal code and tests no longer depend on mutating exported `ToolChoice` package variables.
- A canonical immutable path exists for auto, none, and any tool choices.
- Any retained exported mutable variables are explicitly justified as compatibility debt in docs or comments.
- The three listed validation commands pass.

Outcome:
- Enumerated all `ToolChoiceAuto`, `ToolChoiceNone`, and `ToolChoiceAny` references before editing.
- Added `AutoToolChoice`, `NoneToolChoice`, and `AnyToolChoice` helpers that return fresh `ToolChoice` values.
- Updated provider tests to use the canonical helper path and added coverage that helper results are fresh values.
- Retained the exported mutable variables for stable API compatibility and documented the compatibility debt in code comments and `docs/modules/x-ai/README.md`.
- Re-ran the symbol search; remaining Go references are only the retained compatibility variables in `provider.go`.
- Validation passed:
  - `go test -race -timeout 60s ./x/ai/provider`
  - `go test -timeout 20s ./x/ai/provider`
  - `go vet ./x/ai/provider`
