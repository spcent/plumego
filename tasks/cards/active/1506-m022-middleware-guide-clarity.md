# Card 1506

Milestone: M-022
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P2
State: active
Primary Module: middleware
Owned Files:
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`
Depends On:

Goal:
- Clarify the `middleware` family guidance so agents can distinguish the HTTP
  rate-limit adapter from the reusable resilience primitive.
- Document which constructor naming exceptions are already intentional.

Scope:
- Update `middleware/module.yaml` and the module primer to explain the
  difference between `selection_guide.rate_limiting` and
  `landing_zones.rate_limit`.
- Document the existing constructor-pattern exceptions instead of implying a
  uniform rename target.

Non-goals:
- Do not rename middleware packages or constructors.
- Do not change runtime rate-limit behavior.
- Do not move code between `middleware`, `security`, and `x/resilience`.

Files:
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`

<!-- none; docs/spec clarification card -->

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- `docs/modules/middleware/README.md`

Validation:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `gofmt -l .`

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
<!-- Agent fills this after completion: what changed and why. -->
