# Card 2259

Milestone:
Recipe: specs/change-recipes/review-only.yaml
Priority: P1
State: active
Primary Module: x/rest
Owned Files:
- x/rest/module.yaml
- docs/modules/x-rest/README.md
- docs/ROADMAP.md
- docs/EXTENSION_STABILITY_POLICY.md
Depends On:

Goal:
Evaluate `x/rest` against the extension `experimental` to `beta` promotion policy and prepare the smallest promotion or blocker record.

Scope:
- Check the public `x/rest` entrypoints against the documented beta criteria.
- Verify whether the required two-release API freeze evidence exists.
- If criteria are met, update status and docs consistently.
- If criteria are not met, record the exact blocker in the module primer and roadmap.

Non-goals:
- Do not change `x/rest` runtime behavior.
- Do not add new CRUD features.
- Do not promote unrelated `x/*` packages.

Files:
- `x/rest/module.yaml`
- `docs/modules/x-rest/README.md`
- `docs/ROADMAP.md`
- `docs/EXTENSION_STABILITY_POLICY.md`

Tests:
- `go test -timeout 20s ./x/rest/...`
- `go vet ./x/rest/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required when status or blocker language changes.

Done Definition:
- `x/rest` has either a policy-compliant `beta` promotion patch or an explicit blocker recorded in the live docs.
- The primer, roadmap, stability policy, and manifest agree on status.

Outcome:
