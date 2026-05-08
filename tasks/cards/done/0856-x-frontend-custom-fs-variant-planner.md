# Card 0856

Milestone:
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files: x/frontend/config.go, x/frontend/compression.go, x/frontend/compression_test.go, x/frontend/README.md, docs/modules/x-frontend/README.md, docs/extension-evidence/x-frontend.md
Depends On: 0727

Goal:
Give custom `http.FileSystem` integrations an explicit way to avoid lazy `.br` and `.gz` probing on original responses.

Scope:
- Add an exported precompressed variant planner interface or option that callers can implement for custom filesystems.
- Use the explicit plan to decide variant availability and `Vary: Accept-Encoding` without per-request miss probing.
- Preserve current lazy behavior for custom filesystems that do not provide a plan.
- Add focused tests proving planned custom FS mounts avoid lazy `.br`/`.gz` probes.
- Update docs and evidence so the custom FS performance boundary has a stable mitigation.

Non-goals:
- Do not change directory-backed planning behavior.
- Do not require custom filesystems to implement the new capability.
- Do not add dependencies.

Files:
- `x/frontend/config.go`
- `x/frontend/compression.go`
- `x/frontend/compression_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
- Document the new mitigation for remote/generated/custom filesystems.

Done Definition:
- Custom FS users can opt into a variant plan and avoid lazy miss probes.
- Non-planned custom FS behavior remains unchanged.
- Targeted x/frontend tests and vet pass.

Outcome:
- Added `PrecompressedVariants` and `WithPrecompressedVariantPlan`.
- Used explicit plans for non-directory filesystems to avoid lazy `.br`/`.gz` probes on original responses.
- Preserved lazy probing for custom filesystems without explicit plans.
- Documented the custom FS mitigation in package docs, module docs, and evidence.
- Validation Run:
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
