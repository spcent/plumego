# Card 0771

Milestone:
Recipe: specs/change-recipes/docs.yaml
Priority: P2
State: done
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/config/README.md
- x/data/sharding/module.yaml
- docs/modules/x-data/README.md
Depends On:
- 0770-x-data-kvengine-default-helper-boundary

Goal:
Synchronize sharding config documentation and public API inventory after the stable-readiness fixes.

Scope:
- Update sharding config README for escaped DSN construction and validated env merge behavior.
- Refresh sharding public entrypoint notes for cluster/router ownership and cross-shard semantics.
- Keep examples limited to implemented behavior.

Non-goals:
- Do not change sharding runtime behavior in this docs card.
- Do not mark x/data stable.
- Do not edit unrelated module docs.

Files:
- x/data/sharding/config/README.md
- x/data/sharding/module.yaml
- docs/modules/x-data/README.md

Tests:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow
- go vet ./x/data/sharding/...

Docs Sync:
- This card is the docs sync.

Done Definition:
- Sharding config README matches current behavior.
- Module manifest inventory matches exported stable-readiness surface.
- Checks pass.

Outcome:
- Updated sharding config README for DSN escaping/quoting, env merge validation,
  ClusterDB ownership, and QueryRow default-shard semantics.
- Expanded sharding module public entrypoint inventory for exported routing,
  parser, resolver, strategy, metrics, logging, and health symbols.
- Kept docs limited to implemented behavior.

Validation:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `go vet ./x/data/sharding/...`
