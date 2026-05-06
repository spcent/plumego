# Card 0775

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
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
- Replaced full-file prefiltering in conformance scan discovery with imports-only parsing.
- Added `TestConformanceScanCoverageAndBudget` to ensure critical external paths are visible and the scan remains under a 150-file maintenance budget.
- Documented that conformance scan coverage and budget are self-checked.

Validation:
- go test -run TestConformanceScanCoverageAndBudget -count=1 -timeout 20s ./contract
- go test -timeout 20s ./contract/...
- go vet ./contract/...
