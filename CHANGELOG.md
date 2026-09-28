# Changelog

All notable changes to Plumego are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Detailed per-release notes live under [`docs/release/`](docs/release/).

## [Unreleased]

### Fixed
- README: repaired the `COMPATIBILITY.md` link target (`README.md`).
- `.github/FUNDING.yml`: removed unused template placeholder entries.

### Changed
- CI: stable-root API snapshot comparison and doc-snippet compilation are now
  enforced in the `quality-gates` workflow instead of only locally.
- CI: added CodeQL analysis and `govulncheck` dependency auditing.
- CI: GitHub Actions pinned to immutable SHAs; Dependabot now tracks the
  `github-actions` ecosystem.
- Added a release workflow that attaches `cmd/plumego` binaries to version tags.
- Added `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, this changelog, and GitHub
  issue templates.
- `log`: `*Ctx` logging methods now surface a request ID carried via
  `log.WithRequestID`/`log.RequestIDFromContext`; the `requestid` middleware
  populates it.
- `contract`: `APIError.Details()` deep-copy no longer uses reflection.
- `security/abuse`, `store/cache`, `log`: constructor panic contracts are now
  documented consistently as `MustXxx`-style, with error-returning
  alternatives.
- Migrated remaining `math/rand` uses to `math/rand/v2` (`middleware/requestid`,
  `x/gateway`, `x/data`, `x/messaging`).
- `store`, `log`, `x/tenant`: best-effort cleanup and recovery errors are now
  reported (deferred temp-file removal, rotation/close failures to stderr,
  quota-compensation release errors) instead of being silently discarded.

## [v1.1.0] - 2026-05-18

Post-v1 minor release focused on release-control-plane hardening, agent
workflow discoverability, website documentation parity, and extension evidence
readiness. No intentional stable-root public API changes.

Highlights:
- Agent workflow: route, diff-validation, task-bundle, recipe, and manifest
  guidance expanded.
- Website and docs: release posture, architecture, migration, benchmark, and
  bilingual pages synchronized.
- Extension evidence: first `v1.0.0` release-ref evidence points for selected
  experimental candidates.
- Maintenance: GitHub Actions runtime and CLI install guidance aligned with the
  source-checkout model.

See [`docs/release/v1.1.0.md`](docs/release/v1.1.0.md) for the full notes.

## [v1.0.0] - 2026-05-15

First stable release. Nine GA root packages with frozen v1 APIs, optional
`x/*` extension families, and the agent-first control plane.

Highlights:
- Stable roots: `core`, `router`, `contract`, `middleware`, `security`,
  `store`, `health`, `log`, `metrics`.
- Reference apps and the `cmd/plumego` CLI.
- Boundary checks, stable API snapshots, and coverage gates in CI.

See [`docs/release/v1.0.0-release-notes.md`](docs/release/v1.0.0-release-notes.md)
for the full notes.

## [v0.2.0] - pre-v1

Experimental preview prior to the v1 stabilization effort. See the tagged
history and [`docs/release/`](docs/release/) for details.

[Unreleased]: https://github.com/spcent/plumego/compare/v1.1.0...HEAD
[v1.1.0]: https://github.com/spcent/plumego/releases/tag/v1.1.0
[v1.0.0]: https://github.com/spcent/plumego/releases/tag/v1.0.0
