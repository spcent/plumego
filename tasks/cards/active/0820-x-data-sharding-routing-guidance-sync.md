# Card 0820

Priority: P1
State: active
Primary Module: x/data
Owned Files:
- `docs/modules/x-data/README.md`
- `x/data/sharding/module.yaml`
- `x/data/sharding/config/README.md`
- `x/data/sharding/config/examples/cluster-simple.json`
Depends On:

Goal:
- Clarify `x/data/sharding` strategy selection, routing limits, and config examples without widening the module surface.

Scope:
- Align the x/data primer, sharding manifest, and sharding config README on safe-default cross-shard behavior, transaction entrypoints, and strategy selection guidance.
- Tighten the simplest config example only if the docs need a concrete example to match implemented defaults.
- Keep the guidance explicit about routing limits and safe defaults rather than generic sharding theory.

Non-goals:
- Do not redesign cross-shard policy behavior in this card.
- Do not add new sharding strategies.
- Do not broaden this card into file storage or rw guidance.

Files:
- `docs/modules/x-data/README.md`
- `x/data/sharding/module.yaml`
- `x/data/sharding/config/README.md`
- `x/data/sharding/config/examples/cluster-simple.json`

Tests:
- `go test -timeout 20s ./x/data/sharding/...`
- `go test -race -timeout 60s ./x/data/sharding/...`
- `go vet ./x/data/sharding/...`

Docs Sync:
- Keep sharding routing guidance aligned across the x/data primer, sharding manifest, and sharding config README/examples.

Done Definition:
- Sharding docs state the same safe defaults and routing limits as the implementation.
- The simplest config example matches the documented defaults where it is used.
- Targeted sharding validation stays green.

Outcome:
