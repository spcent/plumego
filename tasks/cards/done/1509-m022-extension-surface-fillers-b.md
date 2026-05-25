# Card 1509

Milestone: M-022
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/gateway
Owned Files:
- `x/gateway/discovery/module.yaml`
- `x/gateway/ipc/module.yaml`
- `x/rpc/client/module.yaml`
- `x/rpc/gateway/module.yaml`
- `x/rpc/server/module.yaml`
Depends On: 1503

Goal:
- Add missing `public_entrypoints` to the subordinate `x/gateway` and `x/rpc`
  manifests.
- Remove the ghost `core/components/**` boundary rule from those manifests.

Scope:
- Touch only the five manifests listed in `Owned Files`.

Non-goals:
- Do not change `x/gateway` or `x/rpc` runtime behavior.
- Do not add root-family docs edits in this card unless a manifest wording fix
  is impossible without them.
- Do not handle the duplicate `ErrHandlerNil` issue here; that belongs to card
  `1513`.

Files:
- `x/gateway/discovery/module.yaml`
- `x/gateway/ipc/module.yaml`
- `x/rpc/client/module.yaml`
- `x/rpc/gateway/module.yaml`
- `x/rpc/server/module.yaml`

Acceptance Tests:
<!-- none; manifest-only card -->

Tests:
- `go test -timeout 20s ./x/gateway/... ./x/rpc/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- None expected.

Validation:
- `go test -timeout 20s ./x/gateway/... ./x/rpc/...`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/public-entrypoints-sync`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Added concrete `public_entrypoints` inventories to the subordinate
  `x/gateway` and `x/rpc` manifests listed in the card.
- Removed the ghost `core/components/**` rule from the `x/gateway`
  subordinate manifests touched here.
- Validation:
  - `go test -timeout 20s ./x/gateway/... ./x/rpc/...`
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/public-entrypoints-sync`
