# Card 1312

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/data/rw
Owned Files:
- x/data/rw/policy.go
- x/data/rw/policy_test.go
- docs/modules/x-data/README.md
Depends On:
- 0768-x-data-file-s3-cleanup-and-key-contract

Goal:
Reduce false primary routing caused by lock keywords inside SQL string literals.

Scope:
- Mask or parse string literals before detecting `FOR UPDATE` and `FOR SHARE`.
- Preserve fail-safe primary routing for actual locking reads.
- Add tests for literals and real lock clauses.

Non-goals:
- Do not introduce a SQL parser dependency.
- Do not change write detection semantics.
- Do not change cluster lifecycle behavior.

Files:
- x/data/rw/policy.go
- x/data/rw/policy_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/rw
- go test -race -timeout 60s ./x/data/rw
- go vet ./x/data/rw

Docs Sync:
- Update x/data docs to describe the refined lock-read heuristic.

Done Definition:
- Literal-only lock keywords no longer force primary routing.
- Real locking reads still route to primary.
- Tests and docs cover the heuristic boundary.

Outcome:
- Masked SQL string literals and comments before checking `FOR UPDATE` and
  `FOR SHARE` lock clauses.
- Added tests for literal/comment false positives and real lock clauses.
- Documented the refined SQLTypePolicy heuristic.

Validation:
- `go test -timeout 20s ./x/data/rw`
- `go test -race -timeout 60s ./x/data/rw`
- `go vet ./x/data/rw`
