# Post-v1 Release Evidence

Use this file as the first release evidence read after `v1.0.0`.

## Published v1

- Final tag: `v1.0.0`
- Tag object: `cde6de1d34c836584c54ba0df86dab3cf92ba6e2`
- Tag target: `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96`
- Final tag GitHub Actions run: `25922384589`
- Final tag result: PASS
- Release notes: `docs/release/v1.0.0.md`
- RC notes: `docs/release/v1.0.0-rc.1.md`
- v1 release execution verify artifact: `tasks/milestones/M-005.verify.md`

Do not rewrite the published tag. Any post-v1 fix belongs in a later patch or
minor release lane.

## Post-v1 Maintenance Lane

- Milestone: `tasks/milestones/active/M-006.md`
- Plan: `tasks/milestones/M-006.plan.md`
- Verify artifact: `tasks/milestones/M-006.verify.md`
- First maintenance CI run after action-runtime cleanup: `25954419567`

M-006 is for generated data, CI upkeep, onboarding truth, release evidence, and
clear bug fixes. It is not a feature lane for new stable-root APIs.

## CLI Install Boundary

`cmd/plumego` is a nested Go module. v1 CLI validation is source-checkout based;
tagged `go install github.com/spcent/plumego/cmd/plumego@<tag>` is not
advertised until a release checklist verifies that exact tag.

## Extension Promotion Boundary

Current beta extension families:

- `x/gateway`
- `x/observability`
- `x/rest`
- `x/websocket`

Remaining candidates must stay blocked until `specs/extension-beta-evidence.yaml`
records two release refs, release-backed API snapshots, and owner sign-off:

- `x/tenant`
- `x/frontend`
- `x/ai` stable-tier subpackages
- selected `x/data` surfaces
- `x/discovery:core-static`
- `x/messaging:app-facing-service`

Run before any extension promotion:

```bash
go run ./internal/checks/extension-beta-evidence
go run ./internal/checks/extension-maturity
```
