# Card 0762

Milestone:
Recipe: specs/change-recipes/refactor-api.yaml
Priority: P1
State: active
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
- x/data/kvengine/module.yaml
- docs/modules/x-data/README.md
Depends On:
- 0761-x-data-stable-readiness-third-gate

Goal:
Remove the ambiguous kvengine auto-detect double-switch so callers can explicitly choose auto-detect behavior before the API is treated as stable.

Scope:
- Replace the `AutoDetectFormat` plus `DisableAutoDetect` double-switch with one clear option.
- Update defaulting so zero-value behavior remains documented and deterministic.
- Update tests and public API inventory for the chosen option.

Non-goals:
- Do not change serializer wire formats.
- Do not change WAL sync semantics.
- Do not move kvengine into stable store.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/kv_test.go
- x/data/kvengine/module.yaml
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/kvengine
- go test -race -timeout 60s ./x/data/kvengine
- go vet ./x/data/kvengine

Docs Sync:
- Update x/data docs to describe the frozen auto-detect option surface.

Done Definition:
- There is one auto-detect configuration knob.
- Explicit disable is distinguishable from zero-value defaults.
- Tests, manifest, and docs reflect the stable option.

