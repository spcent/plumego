# Card 0750

Milestone: v1
Recipe: specs/change-recipes/doc-sync.yaml
Priority: P1
State: done
Primary Module: reference
Owned Files:
- `reference/standard-service/`
- `reference/workerfleet/`
- `cmd/plumego/`
- `docs/release/v1.0.0-rc.1.md`
- `README.md`
Depends On: 0746

Goal:
- Align the canonical reference app and CLI contract with the v1 release position.

Problem:
For v1, users need one canonical wiring path and clear CLI expectations. `reference/standard-service` should be the supported reference, `reference/workerfleet` should not look like a stable v1 surface, and `cmd/plumego` should be documented as a command-line tool rather than an importable API.

Scope:
- Verify `reference/standard-service` still demonstrates stable-root usage.
- Ensure `reference/workerfleet` wording does not imply canonical v1 support.
- Confirm `cmd/plumego` checks are included in local and CI release gates.
- Update release notes and README wording if reference or CLI support is ambiguous.
- Card-split any CLI runtime defect found during validation.

Non-goals:
- Do not redesign scaffold output.
- Do not add new CLI commands.
- Do not turn `reference/workerfleet` into a v1 supported surface.

Files:
- `reference/standard-service/`
- `reference/workerfleet/`
- `cmd/plumego/`
- `docs/release/v1.0.0-rc.1.md`
- `README.md`

Tests:
- `go test ./reference/standard-service/...`
- `cd cmd/plumego && go test ./...`
- `cd cmd/plumego && go vet ./...`

Docs Sync:
- Required for README and release support wording.

Done Definition:
- `reference/standard-service` is the only canonical v1 application wiring example.
- `reference/workerfleet` is clearly non-canonical or blocked where behavior is incomplete.
- CLI validation is part of the release gate evidence.
- Public docs do not imply `cmd/plumego` is a stable Go import surface.

Outcome:
- Verified `reference/standard-service` remains the canonical stable-root-only v1 application reference.
- Updated `reference/workerfleet/README.md` to state it is not the canonical v1 reference application or a reusable stable Plumego surface.
- Verified `cmd/plumego/README.md` documents CLI support as command-line behavior, not a Go import surface.
- Updated release notes with the reference and CLI audit result.
- Validation passed:
  - `go test ./reference/standard-service/...`
  - `cd cmd/plumego && go test ./...`
  - `cd cmd/plumego && go vet ./...`
