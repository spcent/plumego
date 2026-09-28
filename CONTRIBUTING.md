# Contributing to Plumego

Thanks for your interest in Plumego! This project is built as a high-quality,
maintainable open-source sample: the standard library is the centerpiece, and
every change is expected to keep the stable surface small, dependency-free, and
`net/http`-compatible.

This guide covers how to contribute code, report bugs, and work with the
project's agent-first tooling. It applies to both human and AI contributors.

## Code of Conduct

All contributors are expected to follow our [Code of Conduct](CODE_OF_CONDUCT.md).
Be kind, be specific, and assume good faith.

## Getting Started

1. Fork the repository and clone your fork.
2. Install the [Go toolchain](https://go.dev/dl/) — the module requires Go 1.26+.
3. Make a branch: `git checkout -b feat/my-change`.
4. Run the local quality gates before opening a PR: `make gates` (see below).

The repository is a single Go module with no third-party runtime dependencies in
the stable roots. `x/*` extensions and the `reference/*` apps are also part of
the main module. `cmd/plumego`, `examples/*`, and `reference/*` are separate Go
modules that use a local `replace` while developing against this checkout.

## What We Value

- **Stdlib first.** Prefer the standard library over new dependencies. The main
  module is stdlib-only by rule; adding a dependency requires an explicit,
  documented exception.
- **Small stable surface.** `core`, `router`, `contract`, `middleware`,
  `security`, `store`, `health`, `log`, and `metrics` carry a v1 compatibility
  guarantee. Prefer adding new capability under `x/*` over growing a stable root.
- **`net/http` compatibility.** Handlers stay
  `func(http.ResponseWriter, *http.Request)`, middleware wraps `http.Handler`.
- **Tests next to behavior.** Every behavior change ships with focused tests.
- **Docs in sync.** Update README, module primers, and API evidence when you
  change public API, config, security, or lifecycle behavior.

## Reporting Bugs

Open an issue using the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md).
Include:

- Go version and OS.
- A minimal, runnable reproduction (a small `main.go` is ideal).
- Expected vs. actual behavior.

Security issues must **not** be reported in public issues. See
[SECURITY.md](SECURITY.md) for the private disclosure process.

## Requesting Features

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md).
Describe the problem you are solving and a concrete API sketch. Features that
expand the stable root surface are held to a higher bar; consider whether the
capability belongs in an `x/*` extension.

## Development Setup

```bash
make setup-hooks   # installs the pre-push quality-gate hook (optional)
make fmt           # gofmt -w .
make validate-diff # minimal gate profile for your current diff (preferred)
make gates         # full local mirror of CI
make website-gates # docs site content checks + static build (slow)
```

### Quality gates

`make gates` mirrors CI and includes: repo boundary checks, stable API snapshot
comparison for `core`, doc snippet compilation, `go vet ./...`, race tests,
stable-module coverage >= 70%, and `website/src/generated` staleness.

`make validate-diff` auto-selects the minimal gate profile for the current diff
and is the fastest way to check an in-progress change.

## Submitting a Pull Request

1. Fill in the [PR template](.github/pull_request_template.md). It asks for a
   concise summary, the exact scope you touched, public API impact, and the
   verification evidence (test output or CI links).
2. The PR template uses "safe refactor zones" to signal how risky a change is.
   Zone C (API boundary) and Zone D (do not refactor) changes need a design note
   and maintainer sign-off — say so explicitly in the description.
3. CI runs formatting, boundary checks, `go vet`, race tests, coverage gates,
   and the docs-site build. Address failures before requesting review.
4. Keep each PR scoped to one logical change. Small, reviewable PRs land faster.

## Commit Style

Commits follow conventional commits:

- `feat:` new capability
- `fix:` bug fix
- `refactor:` behavior-preserving change
- `perf:` performance improvement
- `docs:` documentation only
- `chore:` tooling, metadata, CI

Keep the subject under 72 characters and add a body when the change needs
context. If your change updates a task card under `tasks/cards/`, mention the
card id in the commit subject.

## Working with the Agent-First Tooling

Plumego is maintained with an agent-first control plane: `docs/` explains the
architecture, `specs/` records machine-checkable boundaries, `tasks/` defines
reviewable execution units, and `reference/` shows canonical wiring.

**This tooling is optional for external contributors.** You are not required to
run `codex`, scaffold milestone specs, or author task cards to contribute. What
matters is that your change passes the gates and follows the boundary rules. If
you are curious about the workflow, read
[docs/concepts/agent-first.md](docs/concepts/agent-first.md).

## Boundary Rules in Brief

- Stable roots must not import `x/*`, `reference/*`, `cmd/*`, or use-case modules.
- The root `github.com/spcent/plumego` facade stays a thin alias surface; do not
  expand it.
- No hidden globals, `init()` registration, route auto-discovery, or
  reflection-based wiring.
- Never log or return secrets. Secret/signature comparisons use timing-safe
  checks.
- Middleware is transport-only; business policy belongs in the app or `x/*`.

The full operational rules live in [AGENTS.md](AGENTS.md) and
[specs/dependency-rules.yaml](specs/dependency-rules.yaml). The canonical style
guide is [docs/reference/canonical-style-guide.md](docs/reference/canonical-style-guide.md).

## Licensing

By contributing, you agree that your contributions are licensed under the
[MIT License](LICENSE).
