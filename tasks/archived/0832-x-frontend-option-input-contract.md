# Card 0832

Milestone:
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files: x/frontend/config.go, x/frontend/response_test.go, x/frontend/README.md, docs/modules/x-frontend/README.md
Depends On:

Goal:
Make x/frontend option input ownership and MIME extension validation stable-grade and unsurprising.

Scope:
- Copy MIME type maps when `WithMIMETypes` is constructed, matching `WithHeaders` input ownership.
- Reject MIME extension keys that cannot be produced by `path.Ext` matching, including separators, whitespace, control characters, and empty normalized extensions.
- Add focused tests for caller map mutation after option construction and invalid MIME extension keys.
- Sync docs so callers know option maps are captured by the exported helper.

Non-goals:
- Do not change MIME value validation beyond the existing header-safety checks.
- Do not add dependencies.
- Do not promote `x/frontend` status.

Files:
- `x/frontend/config.go`
- `x/frontend/response_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
- Update package and module docs for option map capture and MIME extension key constraints.

Done Definition:
- MIME maps are copied before caller mutation can affect a constructed option.
- Invalid MIME extension keys fail mount construction with a focused error.
- Targeted x/frontend tests and vet pass.

Outcome:
- Implemented MIME type map capture at option creation time.
- Added strict single-extension MIME key validation.
- Updated x/frontend docs with option ownership and extension-key rules.
- Validation Run:
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
