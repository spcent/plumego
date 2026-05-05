# Card 0757

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: router
Owned Files: router/static.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0756-router-pool-cache-benchmark-evidence

Goal:
Make local static containment semantics honest and stable around symlink races.

Scope:
- Clarify that local static mounts are intended for read-only or trusted roots.
- Keep the current portable stdlib symlink containment behavior.
- Update comments and docs so router does not promise race-free protection for
  concurrently mutated static roots.

Non-goals:
- Adding platform-specific `openat` or non-stdlib filesystem dependencies.
- Adding frontend asset policy.
- Changing static file serving behavior.

Files:
- router/static.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Static containment is documented as portable best-effort for trusted/read-only
  static roots.
- Code comments no longer imply race-free symlink containment.
- Router vet passes.

Outcome:
