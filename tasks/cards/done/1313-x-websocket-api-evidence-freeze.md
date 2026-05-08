# Card 1313

Milestone:
Recipe: specs/change-recipes/extension-stability.yaml
Priority: P2
State: active
Primary Module: x/websocket
Owned Files: x/websocket/module.yaml, docs/extension-evidence/x-websocket.md, docs/extension-evidence/x-websocket-public-api-inventory.md
Depends On: 0763, 0764, 0765, 0766, 0767, 0768

Goal:
Refresh websocket stable-readiness evidence after the follow-up implementation pass.

Scope:
- Reconcile `module.yaml` public entrypoints with the current exported surface.
- Update evidence blockers and remaining owner-signoff requirements.
- Record release-readiness status honestly without promoting before evidence exists.
- Keep the module experimental if release refs, snapshots, or owner sign-off are still missing.

Non-goals:
- Promote `x/websocket` to stable without evidence.
- Invent release refs or owner approval.
- Change runtime behavior.

Files:
- x/websocket/module.yaml
- docs/extension-evidence/x-websocket.md
- docs/extension-evidence/x-websocket-public-api-inventory.md

Tests:
- go run ./internal/checks/module-manifests
- go test -timeout 20s ./x/websocket/...

Docs Sync:
- Evidence docs are the primary output.

Done Definition:
- Stable-readiness ledger matches current code and known blockers.
- Module manifest check passes.
- Module tests pass.

Outcome:
Refreshed the websocket stable-readiness ledger and current-head API snapshot
after the follow-up implementation pass. Kept `x/websocket` experimental
because release refs, release API snapshots, and `realtime` owner sign-off are
still missing. Verified module manifest checks, snapshot comparison, and module
tests.
