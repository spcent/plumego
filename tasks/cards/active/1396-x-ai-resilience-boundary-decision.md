# Card 1396

Milestone: v1-cleanup-phase-3
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: active
Primary Module: x/ai
Owned Files:
- docs/architecture/x-ai-resilience-boundary.md
- docs/modules/x-ai/README.md
- docs/modules/x-resilience/README.md
- x/ai/module.yaml
- x/resilience/module.yaml
Depends On:
- 1395

Goal:
- Record the v1 boundary decision between `x/ai` resilience wrappers and generic `x/resilience` primitives before any public type migration.

Scope:
- Add an architecture decision document for `x/ai` versus `x/resilience`.
- State that `x/resilience` owns cross-extension primitives such as generic circuit breakers, rate limiters, and middleware adapters.
- State that `x/ai/resilience` owns AI-provider orchestration, provider wrapping, AI request keying, model/provider semantics, and AI error classification.
- Mark `x/ai/circuitbreaker` and `x/ai/ratelimit` as retained compatibility surfaces for current AI wrappers, not destinations for new generic behavior.
- Update module docs/manifests only where they clarify the boundary.

Non-goals:
- Do not migrate public types in this card.
- Do not add compatibility aliases yet.
- Do not move AI provider wrapping into `x/resilience`.
- Do not involve stable roots such as `contract`, `middleware`, or `security`.

Files:
- docs/architecture/x-ai-resilience-boundary.md
- docs/modules/x-ai/README.md
- docs/modules/x-resilience/README.md
- x/ai/module.yaml
- x/resilience/module.yaml

Tests:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/extension-maturity
- go run ./internal/checks/dependency-rules

Docs Sync:
- This is a docs-first boundary card; keep docs and manifests consistent.

Done Definition:
- New generic resilience work has an unambiguous landing zone in `x/resilience`.
- New AI-provider resilience work has an unambiguous landing zone in `x/ai/resilience`.
- The decision explicitly avoids public type migration until a later compatibility card.

Outcome:

