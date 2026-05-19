# Card 0757

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/module.yaml, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0716-router-route-metadata-concurrency

Goal:
Align router's stable control plane with its actual exported surface and documented stable behavior.

Scope:
- Update router/module.yaml public entrypoints to match the stable API snapshot.
- Document stable decisions for ANY, named route collisions, URL parameter behavior, HEAD fallback, and static FS examples where needed.
- Keep active queue state accurate after the router cards are completed.

Non-goals:
- Public API changes.
- New code behavior.
- Moving static asset policy from x/frontend into router.

Files:
- router/module.yaml
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow
- go vet ./router/...

Docs Sync:
- Required for router stable behavior and module manifest.

Done Definition:
- Router docs and module manifest list the same stable public surface.
- StaticFS example is accurate for embedded subdirectories.
- Manifest/workflow checks and router vet pass.

Outcome:
Aligned router/module.yaml and docs/modules/router/README.md with the exported stable router surface. Documented current stable behavior for ANY guidance, named-route collisions, URL parameter handling, HEAD fallback, warm-cache matching, and embedded StaticFS usage.

Validation:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow
- go vet ./router/...
