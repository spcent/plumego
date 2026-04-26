# Card 2272

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P2
State: active
Primary Module: x/ai
Owned Files:
- x/ai/module.yaml
- docs/modules/x-ai/README.md
- docs/ROADMAP.md
- docs/EXTENSION_STABILITY_POLICY.md
Depends On: 2271

Goal:
Turn the `x/ai` stable-tier subpackages into a clearer adoption path without promoting the whole family.

Scope:
- Re-check the manifest-declared stable-tier subpackages: `provider`, `session`, `streaming`, and `tool`.
- Document the recommended AI service path using only stable-tier subpackages.
- Keep orchestration, semantic cache, marketplace, distributed, and resilience workflows experimental unless policy evidence supports a narrower claim.
- Record whether subpackage-level beta evaluation should become follow-up cards.

Non-goals:
- Do not promote the root `x/ai` family to `beta` or `ga`.
- Do not add provider globals or implicit registration.
- Do not add network-dependent tests.

Files:
- `x/ai/module.yaml`
- `docs/modules/x-ai/README.md`
- `docs/ROADMAP.md`
- `docs/EXTENSION_STABILITY_POLICY.md`

Tests:
- `go test -timeout 20s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `go vet ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required when adoption guidance or stability tier wording changes.

Done Definition:
- Users can identify the safe first path for AI service work.
- The root `x/ai` package remains experimental and does not overstate compatibility.

Outcome:
- Re-checked `x/ai/module.yaml`: `provider`, `session`, `streaming`, and `tool`
  remain the stable-tier adoption path while the root `x/ai` family remains
  experimental.
- Documented a recommended service adoption sequence using only stable-tier
  subpackages and explicit application wiring.
- Kept orchestration, semantic cache, marketplace, distributed execution, and
  resilience workflows experimental in the primer, roadmap, and stability policy.
- Recorded that future beta evaluation should be split into per-subpackage
  follow-up cards after release-history evidence is available.

Validations:
- `go test -timeout 20s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `go vet ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `go run ./internal/checks/module-manifests`
