# Card 0737

Milestone: v1
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/bind_helpers.go
- contract/bind_helpers_test.go
- contract/write_bind_error_test.go
- docs/modules/contract/README.md
Depends On:
- 0736

Goal:
Classify binding programmer/configuration errors as server errors while preserving client classification for malformed request input.

Scope:
- Keep malformed JSON, empty body, extra data, body too large, and invalid query values as client-side bind errors.
- Map invalid bind destinations, nil `Ctx`, nil request, and invalid bind options to internal/server errors.
- Preserve `ErrValidationConfig` as server error behavior.
- Add focused regression tests for the classification matrix.
- Document the bind error classification split.

Non-goals:
- Do not redesign `Ctx` or binding APIs.
- Do not change field validation failures from 400 validation responses.
- Do not add new dependencies or validation rules.

Files:
- contract/bind_helpers.go
- contract/bind_helpers_test.go
- contract/write_bind_error_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests

Docs Sync:
- Update `docs/modules/contract/README.md` for the implemented classification matrix.

Done Definition:
- Programmer/configuration bind failures fail closed as 500/internal metadata.
- Client request input failures remain 4xx with request-specific codes.
- Targeted tests, vet, and manifest checks pass.

Outcome:
- Mapped invalid bind destinations, nil `Ctx`, nil request, and invalid bind options to 500/internal metadata.
- Preserved 4xx classification for malformed JSON, empty body, extra JSON data, oversized body, and invalid query values.
- Added regression coverage for client-input versus programmer/configuration bind classification.
- Documented the classification split in the contract module README.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests
