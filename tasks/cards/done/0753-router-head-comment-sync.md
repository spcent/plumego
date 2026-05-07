# Card 0753

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: done
Primary Module: router
Owned Files: router/dispatch.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0752-router-introspection-empty-slices

Goal:
Sync HEAD suppression comments with current matched-HEAD behavior.

Scope:
- Update `noBodyWriter` comments to describe all matched HEAD routes, including
  explicit HEAD and `MethodAny`.
- Ensure docs remain consistent with the comment.

Non-goals:
- Changing HEAD behavior.
- Adding new tests for already-covered HEAD paths.
- Changing 405 behavior.

Files:
- router/dispatch.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go vet ./router/...

Docs Sync:
- Only if docs require wording adjustment.

Done Definition:
- Comments no longer describe old GET-fallback-only behavior.
- Router vet passes.

Outcome:
- Updated `noBodyWriter` comments to describe all matched HEAD requests,
  including explicit HEAD routes, `MethodAny` fallback, and GET fallback.

Validation:
- go vet ./router/...
