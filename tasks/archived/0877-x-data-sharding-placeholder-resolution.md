# Card 0877

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/resolver.go
- x/data/sharding/resolver_test.go
Depends On:
- 0729-x-data-sharding-queryrow-safe-default

Goal:
Make sharding placeholder indexing deterministic for both ? and PostgreSQL-style numbered placeholders.

Scope:
- Count positional placeholders in conditions for both ? and $N forms.
- Resolve shard key argument indexes correctly when earlier non-shard predicates or UPDATE SET placeholders exist.
- Add focused tests for SELECT and UPDATE statements using PostgreSQL placeholders.

Non-goals:
- Do not introduce a full SQL parser.
- Do not support named placeholders in this card.
- Do not change strategy behavior.

Files:
- x/data/sharding/resolver.go
- x/data/sharding/resolver_test.go

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Not required unless public SQL support documentation changes.

Done Definition:
- `$N` placeholders do not cause shard-key argument misrouting.
- Existing `?` placeholder behavior remains covered.
- sharding normal and race tests pass.

Outcome:
- Counted both `?` and PostgreSQL `$N` placeholders when walking WHERE and UPDATE SET clauses.
- Used explicit `$N` indexes directly for shard-key argument lookup to avoid condition-order misrouting.
- Preserved existing `?` offset behavior for positional placeholders.
- Added SELECT and UPDATE regression tests for PostgreSQL placeholders.

Validation:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding
