# Card 1506

Milestone: M-022
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P2
State: done
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
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Clarified that stable HTTP rate limiting belongs to `middleware/ratelimit`
  while reusable limiter primitives belong to `x/resilience/ratelimit`.
- Documented the intentional constructor naming exceptions for `auth`,
  `cors.StrictDefaultOptions`, and `ratelimit.NewAbuseGuard(...).Middleware()`.
- Validation:
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/agent-workflow`
  - `gofmt -l .`
