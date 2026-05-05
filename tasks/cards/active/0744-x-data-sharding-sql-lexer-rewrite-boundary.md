# Card 0744

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/rewriter.go
- x/data/sharding/rewriter_test.go
- docs/modules/x-data/README.md
Depends On:
- 0743-x-data-stable-readiness-final-gate

Goal:
Make SQL table rewriting avoid string/comment replacement and fail closed for schema-qualified sharding targets.

Scope:
- Replace table names only in SQL code, not inside quoted string literals or SQL comments.
- Preserve existing simple single-statement support.
- Reject schema-qualified logical table targets until parser support is explicit.
- Add tests for strings, line comments, block comments, and schema-qualified input.

Non-goals:
- Do not add a third-party SQL parser.
- Do not implement full SQL grammar support.
- Do not change shard key resolution in this card.

Files:
- x/data/sharding/rewriter.go
- x/data/sharding/rewriter_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Update x/data docs for rewrite limits around strings, comments, and schema-qualified tables.

Done Definition:
- Rewriter does not alter literals or comments containing logical table names.
- Schema-qualified sharding targets fail closed with ErrUnsafeSQLRewrite.
- Existing simple rewrite tests continue to pass.

Outcome:
