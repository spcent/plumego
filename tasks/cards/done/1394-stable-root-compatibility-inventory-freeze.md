# Card 1394

Milestone: M-004
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: done
Primary Module: contract
Owned Files:
- specs/deprecation-inventory.yaml
- docs/stable-api/README.md
- docs/modules/contract/README.md
- docs/modules/security/README.md
- docs/modules/store/README.md
Depends On:
- M-002

Goal:
Freeze the v1 stable-root compatibility inventory before code cleanup starts.

Scope:
- Review stable-root compatibility aliases, lenient constructors, panic-compatible helpers, and no-error compatibility methods.
- Record whether each compatibility path is kept for v1, migrated later, or blocked pending a symbol-change card.
- Align stable-root docs with the compatibility inventory and stable API snapshot policy.

Non-goals:
- Do not remove or rename exported symbols.
- Do not change runtime behavior.
- Do not edit `x/*` extension docs or statuses.

Files:
- specs/deprecation-inventory.yaml
- docs/stable-api/README.md
- docs/modules/contract/README.md
- docs/modules/security/README.md
- docs/modules/store/README.md

Tests:
- go run ./internal/checks/deprecation-inventory -strict
- go run ./internal/checks/module-manifests
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required for stable-root compatibility and v1 freeze wording.

Done Definition:
- Every retained stable-root compatibility path has an explicit v1 keep/migrate decision.
- Stable-root docs and checked inventory do not disagree.
- Target checks pass.

Outcome:
- Recorded retained stable-root compatibility aliases, lenient paths, and
  no-error compatibility methods in the deprecation inventory and stable API
  docs.
- Registered previously untracked workerfleet reference compatibility markers so
  strict inventory checks stay deterministic.
- No runtime behavior or public API changed.

Validation:
- go run ./internal/checks/deprecation-inventory -strict
- go run ./internal/checks/module-manifests
- go run ./internal/checks/dependency-rules
