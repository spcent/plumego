# Card 2062

Milestone: M-024
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/ai
Owned Files:
- `x/ai/provider/module.yaml`
- `x/ai/session/module.yaml`
- `x/ai/module.yaml`
- `docs/modules/x/ai/README.md`
- `docs/EXTENSION_MATURITY.md`
Depends On: M-023

## Goal

Add explicit manifests for `x/ai/provider` and `x/ai/session` so stable-tier
AI metadata is machine-readable at the same granularity as other beta-oriented
subpackage surfaces.

## Scope

Create subpackage manifests, keep `x/ai/module.yaml` as the family index, and
sync the dashboard and primer wording to reference the new manifest authority.

## Non-goals

- Do not promote root `x/ai` to beta or GA.
- Do not change provider or session runtime behavior.
- Do not add manifests for experimental AI subpackages in this card.

## Files

- `x/ai/provider/module.yaml`
- `x/ai/session/module.yaml`
- `x/ai/module.yaml`
- `docs/modules/x/ai/README.md`
- `docs/EXTENSION_MATURITY.md`

## Acceptance Tests

- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/extension-maturity`

## Tests

- `go test -timeout 20s ./x/ai/provider/... ./x/ai/session/...`

## Docs Sync

- `docs/modules/x/ai/README.md`
- `docs/EXTENSION_MATURITY.md`

## Validation

- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/extension-maturity`
- `gofmt -l .`

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

Added explicit `x/ai/provider/module.yaml` and `x/ai/session/module.yaml`
manifests, kept `x/ai/module.yaml` as the family tier index, and updated the
AI primer and extension maturity dashboard to point package-level metadata at
the new subpackage manifests.
