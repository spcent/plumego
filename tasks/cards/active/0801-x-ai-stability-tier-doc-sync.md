# Card 0801

Priority: P1
State: active
Primary Module: x/ai
Owned Files:
- `docs/modules/x-ai/README.md`
- `x/ai/doc.go`
- `x/ai/module.yaml`
Depends On:

Goal:
- Align the repo-native `x/ai` docs with the manifest-defined stable and experimental subpackage tiers.

Scope:
- Update `docs/modules/x-ai/README.md` so it mirrors the exact stability split already declared in `x/ai/module.yaml`.
- Tighten `x/ai/doc.go` if its package-level prose blurs module-wide experimental status or stable-tier subpackage boundaries.
- Keep Phase 8 roadmap language grounded in implemented behavior rather than speculative guarantees.

Non-goals:
- Do not change any `x/ai` API surface in this card.
- Do not add runnable examples in this card.
- Do not promote `x/ai` from experimental module status.

Files:
- `docs/modules/x-ai/README.md`
- `x/ai/doc.go`
- `x/ai/module.yaml`

Tests:
- `go run ./internal/checks/module-manifests`
- `go test -timeout 20s ./x/ai/...`
- `go vet ./x/ai/...`

Docs Sync:
- Keep `docs/modules/x-ai/README.md` and `x/ai/doc.go` aligned with `x/ai/module.yaml` and `docs/ROADMAP.md` Phase 8 language.

Done Definition:
- The `x/ai` primer describes the same stable-tier and experimental-tier split as the manifest.
- Package-level docs do not overstate module-wide stability.
- No stale prose suggests hidden registration, implicit globals, or stable-root promotion.

Outcome:
