# Card 0775

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/conformance_test.go
- docs/modules/contract/README.md
Depends On:
- 0774

Goal:
Add self-checks for contract conformance scan coverage and performance budget risk.

Scope:
- Add a conformance scan self-test that proves expected roots/callsites are visible.
- Add a conservative budget check for the number of files importing `contract`.
- Document that conformance is a maintained stable gate.

Non-goals:
- Do not replace the scanner with a new external dependency.
- Do not add generated baselines.
- Do not change public API behavior.

Files:
- contract/conformance_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update conformance maintenance notes.

Done Definition:
- Scanner coverage and file-count budget regressions fail clearly.
- Target checks pass.

Outcome:

