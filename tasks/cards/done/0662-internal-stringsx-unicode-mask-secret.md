# Card 0662

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/stringsx
Owned Files: internal/stringsx/string.go, internal/stringsx/string_test.go
Depends On: 0661

Goal:
Make `MaskSecret` preserve valid UTF-8 for multibyte secrets.

Scope:
- Count and slice secrets by rune instead of byte.
- Preserve the existing masking rules for ASCII strings.
- Add focused tests for multibyte and short multibyte secrets.

Non-goals:
- Do not change trimming behavior.
- Do not add new string utilities.

Files:
- internal/stringsx/string.go
- internal/stringsx/string_test.go

Tests:
- go test -timeout 20s ./internal/stringsx
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; internal correctness fix.

Done Definition:
- Multibyte secrets are masked without invalid UTF-8 output.
- Existing ASCII masking tests continue to pass.

Outcome:
Completed. `MaskSecret` now preserves and masks by rune so multibyte secrets remain valid UTF-8.

Validation:
- go test -timeout 20s ./internal/stringsx
- go test -timeout 20s ./internal/...
- go vet ./internal/...
