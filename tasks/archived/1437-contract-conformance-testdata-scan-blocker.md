# Card 1437

Milestone: M-005
Recipe: specs/change-recipes/single-module-behavior.yaml
Priority: P0
State: done
Primary Module: contract
Owned Files:
- `contract/conformance_test.go`
- `tasks/cards/done/1437-contract-conformance-testdata-scan-blocker.md`
Depends On:
- 1436

Goal:
- Remove the final `make gates` race between contract conformance scanning and
  `x/data/kvengine` runtime test directories.

Scope:
- Keep the conformance scan source-focused by skipping `testdata` and
  `testdata_*` directories.
- Do not change contract runtime behavior or public API.

Non-goals:
- Do not change `x/data/kvengine` behavior.
- Do not relax contract conformance rules for checked-in source files.

Files:
- `contract/conformance_test.go`

Tests:
- `go test -race -timeout 60s ./contract`
- `GOCACHE=/private/tmp/plumego-gocache GOMODCACHE=/private/tmp/plumego-gomodcache make gates`

Docs Sync:
- Not required; test-only scanner behavior.

Done Definition:
- Contract conformance tests pass under race mode.
- Full release gate can proceed without the transient `x/data/kvengine/testdata`
  directory error.

Outcome:
- `contract` conformance scanning now skips runtime test data directories.
