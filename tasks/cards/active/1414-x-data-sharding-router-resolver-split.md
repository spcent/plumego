# Card 1414

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/router.go
- x/data/sharding/router_plan.go
- x/data/sharding/resolver.go
- x/data/sharding/resolver_rules.go
- x/data/sharding/router_test.go
Depends On:
- 1413

Goal:
- Reduce `x/data/sharding` router and resolver edit radius without changing routing decisions.

Scope:
- Move route planning helpers into `router_plan.go`.
- Move resolver rule matching helpers into `resolver_rules.go`.
- Preserve shard selection, unsupported SQL fallback, metrics, and logging behavior.

Non-goals:
- Do not change parser behavior.
- Do not change strategy implementations.
- Do not change config watcher behavior.

Files:
- x/data/sharding/router.go
- x/data/sharding/router_plan.go
- x/data/sharding/resolver.go
- x/data/sharding/resolver_rules.go
- x/data/sharding/router_test.go

Tests:
- go test -timeout 30s ./x/data/sharding/...
- go vet ./x/data/sharding/...
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless public sharding comments move.

Done Definition:
- Routing plan and resolver rule logic have separate file ownership.
- Existing sharding tests pass.
- No route or shard selection behavior changes.

Outcome:

