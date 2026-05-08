# Card 0929

Milestone: v1
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/trace.go
- contract/trace_test.go
- docs/modules/contract/README.md
- contract/module.yaml
Depends On:
- 0734

Goal:
Expose small validity helpers on `TraceContext` so callers can distinguish partial carriers from complete valid trace context.

Scope:
- Add methods that report whether trace ID and span ID fields are present and valid.
- Add a method that reports whether both required trace identifiers are valid.
- Preserve current storage behavior: `WithTraceContext` remains a defensive carrier and does not reject partial values.
- Add focused tests for full, partial, malformed, and empty trace contexts.
- Update public surface docs and module manifest.

Non-goals:
- Do not add trace header extraction or injection.
- Do not move tracing infrastructure into `contract`.
- Do not change `WithSpanIDString` behavior.

Files:
- contract/trace.go
- contract/trace_test.go
- docs/modules/contract/README.md
- contract/module.yaml

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests

Docs Sync:
- Sync trace helper methods in `contract/module.yaml` and `docs/modules/contract/README.md`.

Done Definition:
- Callers can check trace ID validity, span ID validity, and full context validity without duplicating parse logic.
- Existing trace carrier behavior remains unchanged.
- Targeted tests, vet, and manifest checks pass.

Outcome:
- Added `TraceContext.HasTraceID`, `TraceContext.HasSpanID`, and `TraceContext.Valid`.
- Added focused tests for full, partial, malformed, and empty trace contexts.
- Updated contract public-surface docs and module manifest.
- Preserved existing `WithTraceContext` carrier behavior.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests
