# Card 1012

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/module.yaml
- docs/modules/contract/README.md
Depends On:
- 0741

Goal:
Classify the broad `contract` public surface into stable, compatibility, and guardrail-only groups so v1 support obligations are explicit.

Scope:
- Add a public surface freeze matrix for the current exported entrypoints.
- Mark compatibility carriers that remain supported but are not expansion points.
- Record candidate-removal expectations as future breaking work only.
- Keep behavior unchanged.

Non-goals:
- Do not remove, rename, or deprecate exported symbols.
- Do not add runtime checks.
- Do not change error, response, binding, validation, or context behavior.

Files:
- contract/module.yaml
- docs/modules/contract/README.md

Tests:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow

Docs Sync:
- This is a control-plane and module documentation sync.

Done Definition:
- The contract stable surface has an explicit classification matrix.
- Compatibility helpers are named as retained support surfaces, not preferred expansion paths.
- Manifest checks pass.

Outcome:
- Added a public surface stability matrix to `contract/module.yaml`.
- Documented core stable, request metadata carrier, compatibility helper, and guardrail constant groups in the module README.
- Recorded that narrowing `Ctx`, binding helpers, `ValidateStruct`, exported `APIError` fields, or context carriers is future breaking work only.

Validation:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow
