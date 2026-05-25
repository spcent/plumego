# Card 1505

Milestone: M-022
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/module.yaml`
Depends On:

Goal:
- Replace the wildcard sentinel entry in `contract/module.yaml` with concrete
  exported sentinel names.
- Align the manifest wording with the actual nil-check and write-path usage.

Scope:
- Touch only the `contract` manifest public entrypoint inventory and adjacent
  wording needed to distinguish transport write errors from construction-time
  nil checks.

Non-goals:
- Do not change `contract` code or error values in this card.
- Do not add new error helper families.
- Do not widen this into `x/rpc` duplicate-error cleanup; that belongs to a
  later card.

Files:
- `contract/module.yaml`

<!-- none; manifest-only card -->

Tests:
- `go test -timeout 20s ./contract/...`
- `go run ./internal/checks/public-entrypoints-sync`

Docs Sync:
- `docs/modules/contract/README.md` only if manifest wording becomes inconsistent with current docs.

Validation:
- `go test -timeout 20s ./contract/...`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/public-entrypoints-sync`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Replaced the wildcard sentinel description with the concrete exported errors
  `ErrHandlerNil` and `ErrResponseWriterNil`.
- Updated `guardrail_constants` to use the same exact names so manifest
  consumers can match them without pattern inference.
- Validation:
  - `go test -timeout 20s ./contract/...`
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/public-entrypoints-sync`
