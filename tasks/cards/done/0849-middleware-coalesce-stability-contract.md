# Card 0849

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/coalesce/coalesce.go
- docs/modules/middleware/README.md
- middleware/module.yaml
- docs/stable-api/snapshots/middleware-head.snapshot
Depends On:
- 0727-middleware-observability-wiring-contract

Goal:
Decide and document the stable maturity contract for `middleware/coalesce`.

Scope:
- Keep coalesce as GA stable middleware only for bounded, safe, non-streaming
  transport responses.
- Document its high-risk boundaries: in-flight state, key collision behavior,
  no streaming/SSE/websocket use, and explicit `KeyFunc` requirement for
  sensitive variants.
- Align package comments and module docs with the GA/high-risk contract.
- Ensure stable API snapshot remains intentionally aligned.

Non-goals:
- Do not move coalesce out of the stable root.
- Do not change the public API or default key algorithm.
- Do not add cache freshness or business policy behavior.

Files:
- middleware/coalesce/coalesce.go
- docs/modules/middleware/README.md
- middleware/module.yaml
- docs/stable-api/snapshots/middleware-head.snapshot

Tests:
- go test -timeout 20s ./middleware/coalesce
- go test -timeout 20s ./middleware/conformance

Docs Sync:
- docs/modules/middleware/README.md
- middleware/module.yaml

Done Definition:
- Coalesce has an explicit GA/high-risk stable contract.
- Docs state when it is appropriate and when it must be avoided.
- Targeted tests and conformance checks pass.

Outcome:
- Kept coalesce in the stable root as GA middleware, but documented it as a
  high-risk transport primitive with bounded, safe, non-streaming scope.
- Clarified package and module docs around in-flight state, variant key
  completeness, key collision boundaries, and non-streaming use.
- Added module manifest risk and agent hint coverage for coalesce key safety.
- Stable API snapshot remains unchanged because no public API changed.

Validation:
- `go test -timeout 20s ./middleware/coalesce`
- `go test -timeout 20s ./middleware/conformance`
