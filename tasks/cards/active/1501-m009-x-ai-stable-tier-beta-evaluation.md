# Card 1501

Milestone: M-009
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: active
Primary Module: x/ai
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/evidence/extension/x-ai.md`
- `docs/concepts/extension-maturity.md`
- `x/ai/module.yaml`

Goal:
- Evaluate the four stable-tier x/ai subpackages (provider, session, streaming, tool)
  individually for beta promotion based on release-history evidence from v1.0.0 → v1.1.0.
- For each subpackage that meets beta criteria, update evidence, manifests, and dashboard.
- For subpackages that do not meet criteria, record explicit blockers.

Problem:
x/ai/module.yaml already separates stable-tier subpackages (provider, session, streaming,
tool) from experimental ones (orchestration, semanticcache, marketplace, distributed,
resilience). The stable-tier subpackages have contract tests and signed ledger entries,
but formal beta promotion requires release-history evidence.

Beta criteria per subpackage:
1. Two consecutive minor releases without exported API changes.
2. Checked-in API snapshot for each release ref.
3. Owner sign-off in `specs/extension-beta-evidence.yaml`.
4. Subpackage tests pass under race detector.

Scope:
- Evaluate each of provider, session, streaming, tool independently.
- Run `go run ./internal/checks/extension-release-evidence` for each subpackage
  path once the second qualifying release ref is available.
- If criteria are met for a subpackage: update its evidence ledger entry and the
  maturity dashboard. Do not batch all four into a single update — evaluate individually.
- If criteria are not yet met: record the missing release ref as the explicit blocker.

Non-goals:
- Do not evaluate or promote orchestration, semanticcache, marketplace, distributed,
  or resilience subpackages from this card — they remain explicitly experimental.
- Do not promote x/ai at the module level; subpackages are evaluated individually.
- Do not change any AI provider adapter or session behavior from this card.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/evidence/extension/x-ai.md`
- `docs/concepts/extension-maturity.md`
- `x/ai/module.yaml`
- `docs/modules/x/ai/README.md`
- `docs/release/roadmap.md`

Tests:
- `go test -race -timeout 60s ./x/ai/provider/...`
- `go test -race -timeout 60s ./x/ai/session/...`
- `go test -race -timeout 60s ./x/ai/streaming/...`
- `go test -race -timeout 60s ./x/ai/tool/...`
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- Required per promoted subpackage: update `docs/concepts/extension-maturity.md` selected-beta-surfaces
  table, `x/ai/module.yaml` subpackage status, `docs/modules/x/ai/README.md` status table,
  and `docs/release/roadmap.md` Phase 8 next-work section.

Done Definition:
- Each of the four stable-tier subpackages either has beta status in the evidence ledger
  with snapshots and owner sign-off, OR has explicit blockers documented with unblock
  conditions (minimum: the second qualifying release ref must exist).

Notes:
- The stable-tier subpackages already have deepened contract tests per the Phase 8 roadmap
  entry; this card is about collecting release-history evidence, not adding new tests.
- Provider adapters (ClaudeProvider, OpenAIProvider) are tested offline via httptest;
  confirm their tests pass before evaluation.
