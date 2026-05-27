# Extension Beta Promotion Card Template

Use this template when filing a pull request to promote an `x/*` module from
`experimental` to `beta`. All evidence must be complete before merging.

Copy the template below into the PR description. Replace every `<placeholder>`
with actual values. Do not leave any section empty.

---

## Promotion: `x/<family>` → beta

**Owner:** `<owner>`
**Release refs:** `<older-ref>` → `<newer-ref>`
**Evidence doc:** `docs/extension-evidence/x-<family>.md`

### Evidence Summary

| Criterion | State |
| --- | --- |
| Two release refs recorded | ✓ `<older-ref>`, `<newer-ref>` |
| API unchanged between refs | ✓ confirmed by `extension-release-evidence` |
| Release-backed snapshots on record | ✓ `docs/extension-evidence/snapshots/x-<family>/` |
| Behavior coverage complete | ✓ (list key test files) |
| Primer documents current behavior | ✓ `docs/modules/x/<family>/README.md` |
| Owner sign-off | ✓ recorded in evidence doc |

### Snapshot Comparison Output

Paste the output of:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/<family>/... \
  -base <older-ref> \
  -head <newer-ref> \
  -out-dir docs/extension-evidence/snapshots/x-<family>
```

```
module  ./x/<family>/...
base    <older-ref>
head    <newer-ref>
base_snapshot   docs/extension-evidence/snapshots/x-<family>/base.snapshot
head_snapshot   docs/extension-evidence/snapshots/x-<family>/head.snapshot
api     unchanged
snapshots match
```

### Gate Output

Paste the output of:

```bash
make gates
```

```
(paste full output — must show all gates passing)
```

### Files Changed

- `x/<family>/module.yaml` — `status: experimental` → `status: beta`
- `docs/extension-evidence/x-<family>.md` — added Release Evidence and Owner Sign-Off sections, set Evidence state to complete
- `specs/extension-beta-evidence.yaml` — updated `current_status`, `release_refs`, `api_snapshots`, `owner_signoff`, `blockers`
- `docs/extension-evidence/snapshots/x-<family>/base.snapshot` — new
- `docs/extension-evidence/snapshots/x-<family>/head.snapshot` — new
- `docs/EXTENSION_MATURITY.md` — updated dashboard row
- `README.md` — updated support matrix
- `README_CN.md` — updated support matrix

### Owner Sign-Off

> I confirm that `x/<family>` meets the beta criteria in
> docs/EXTENSION_STABILITY_POLICY.md and accept the beta compatibility
> obligations for the documented `x/<family>` public surface.

— `<owner>`, `<date>`

### Pre-Merge Checklist

- [ ] `go run ./internal/checks/extension-beta-evidence` passes with no violations
- [ ] `go run ./internal/checks/extension-maturity` passes with no violations
- [ ] `make gates` passes
- [ ] `cd website && pnpm sync && pnpm build` passes
- [ ] No stale blocker entries remain in `specs/extension-beta-evidence.yaml`
- [ ] Evidence doc `Evidence state` is `complete`
- [ ] `module.yaml` `status` is `beta`
- [ ] Dashboard row in `EXTENSION_MATURITY.md` shows `beta`
