# Card 2063

Milestone: M-024
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/ai
Owned Files:
- `x/ai/streaming/module.yaml`
- `x/ai/tool/module.yaml`
- `x/ai/module.yaml`
- `docs/modules/x/ai/README.md`
- `docs/EXTENSION_MATURITY.md`
Depends On: 2062

## Goal

Add explicit manifests for `x/ai/streaming` and `x/ai/tool` and align the
stable-tier AI dashboard wording with the new subpackage-level authority.

## Scope

Create the two remaining stable-tier subpackage manifests and update root-family
docs so the stable-tier evidence no longer relies on root `x/ai` prose alone.

## Non-goals

- Do not promote root `x/ai` to beta or GA.
- Do not change streaming or tool runtime behavior.
- Do not add manifests for experimental AI subpackages in this card.

## Files

- `x/ai/streaming/module.yaml`
- `x/ai/tool/module.yaml`
- `x/ai/module.yaml`
- `docs/modules/x/ai/README.md`
- `docs/EXTENSION_MATURITY.md`

## Acceptance Tests

- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/extension-maturity`

## Tests

- `go test -timeout 20s ./x/ai/streaming/... ./x/ai/tool/...`

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

Added explicit `x/ai/streaming/module.yaml` and `x/ai/tool/module.yaml`
manifests, completed stable-tier package-level manifest coverage for `x/ai`,
and updated the family primer and maturity dashboard so stable-tier metadata no
longer relies on root-family prose alone.
