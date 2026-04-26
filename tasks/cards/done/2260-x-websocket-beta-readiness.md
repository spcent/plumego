# Card 2260

Milestone:
Recipe: specs/change-recipes/review-only.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- x/websocket/module.yaml
- docs/modules/x-websocket/README.md
- docs/ROADMAP.md
- docs/EXTENSION_STABILITY_POLICY.md
Depends On: 2259

Goal:
Evaluate `x/websocket` for `beta` readiness and record the promotion decision in the control plane.

Scope:
- Compare public hub, connection, security, and handler entrypoints with the beta criteria.
- Verify that lifecycle, shutdown, capacity, and negative-path tests cover documented behavior.
- If criteria are met, promote the module status and update docs.
- If criteria are not met, record the remaining blocker precisely.

Non-goals:
- Do not change WebSocket protocol behavior.
- Do not add new room or broker features.
- Do not introduce package-global metrics or logging.

Files:
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`
- `docs/ROADMAP.md`
- `docs/EXTENSION_STABILITY_POLICY.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required when status or blocker language changes.

Done Definition:
- `x/websocket` status is either promoted with policy evidence or left experimental with a clear blocker.
- Docs and manifest agree on the decision.

Outcome:
Completed. `x/websocket` remains `experimental` because the repository has no
verifiable two-minor-release API freeze evidence for exported `x/websocket`
symbols. The module primer, roadmap, and extension stability policy now record
that blocker while preserving the existing coverage and boundary evidence.
