# Plumego v1 Release Runbook

Status: Active  
Last updated: 2026-03-10

## 1. Local Preflight

Run the aggregate check from repo root:

```bash
bash scripts/v1-release-readiness-check.sh
```

Optional (include race tests):

```bash
RUN_RACE=1 bash scripts/v1-release-readiness-check.sh
```

## 2. CI Gates (must be green)

Workflow: `.github/workflows/quality-gates.yml`

Required jobs for release candidate/final release:

- `go-quality-gates`
- `go-submodule-quality-gates`
- `v1-release-readiness`

## 3. Branch/Tag Trigger Policy

`quality-gates` runs on:

- Pull requests
- Push to `main`
- Push to `release/*`
- Push tags `v*`
- Manual trigger (`workflow_dispatch`)

## 4. GitHub Required Checks Setup

In repository settings, configure branch protection for `main` and release branches:

1. Enable required status checks.
2. Add required checks:
   - `go-quality-gates`
   - `go-submodule-quality-gates`
   - `v1-release-readiness`
3. Require branch to be up to date before merging.
4. Restrict direct pushes (recommended).

## 5. Release Command Sequence (example)

```bash
# 1) Ensure working tree is clean and checks pass
bash scripts/v1-release-readiness-check.sh

# 2) Create annotated release tag
git tag -a v1.0.0 -m "Plumego v1.0.0"
git push origin v1.0.0
```

## 6. Scope Guard

- GA compatibility promises apply to: `core`, `router`, `middleware`, `contract`, `security`, `store`.
- `tenant/*` and `net/mq/*` remain experimental in v1.0.
