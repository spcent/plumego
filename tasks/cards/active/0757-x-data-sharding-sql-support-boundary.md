# Card 0757

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/parser.go
- x/data/sharding/resolver.go
- x/data/sharding/parser_test.go
- x/data/sharding/resolver_test.go
- docs/modules/x-data/README.md
Depends On:
- 0756-x-data-sharding-config-validation-and-dsn

Goal:
Make the current regex SQL parser fail closed for unsupported shapes instead of pretending to understand broad SQL.

Scope:
- Reject multi-statement input, schema-qualified routing targets, joins, subqueries, and INSERT values shapes that cannot be mapped safely to args.
- Keep supported single-table SELECT/INSERT/UPDATE/DELETE forms working.
- Add negative tests for common misrouting shapes.
- Document the intentionally narrow SQL support boundary.

Non-goals:
- Do not add a SQL parser dependency.
- Do not support joins or arbitrary expressions.
- Do not change SQL rewriting beyond routing support validation.

Files:
- x/data/sharding/parser.go
- x/data/sharding/resolver.go
- x/data/sharding/parser_test.go
- x/data/sharding/resolver_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Record the supported SQL subset and fail-closed behavior.

Done Definition:
- Unsupported SQL shapes return ErrUnsupportedSQL or routing errors.
- Existing supported resolver tests continue to pass.
- New tests cover the stable boundary.
