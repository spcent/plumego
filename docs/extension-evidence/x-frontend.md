# x/frontend Beta Evidence

Module: `x/frontend`

Owner: `frontend`

Current status: `experimental`

Candidate status: `beta`

Evidence state: incomplete

## Current Coverage

- Mount construction coverage includes directory-backed mounts,
  caller-provided `http.FileSystem` mounts, explicit `Mount` registration, nil
  registrar/filesystem handling, duplicate route preflight, missing
  directory/index startup failures, `http.Dir` safety convergence, and
  relative path stability after working-directory changes.
- Path safety coverage includes traversal, encoded traversal, backslash
  traversal, dotted filenames, unsafe backend-open prevention, and directory
  symlink escape rejection for both `RegisterFromDir` and `RegisterFS` with
  `http.Dir` inputs.
- Response semantics coverage includes navigation-only SPA fallback, missing
  asset 404 behavior, HEAD, method restrictions, cache-control split, custom
  pages, MIME overrides, unsafe custom header rejection, and custom page cache
  isolation.
- Precompressed response coverage includes `.br` and `.gz` selection, quality
  ordering, wildcard handling, invalid quality values, `identity` refusal,
  orphan variant rejection, directory variant plans, directory scan error
  fail-fast behavior, `http.Dir` directory-plan behavior, and lazy probing for
  non-`http.Dir` custom filesystems, including lazy `Vary` probing misses and
  stat-error fallback from an unusable compressed variant to the original asset.
- Response error coverage includes original-file stat failures and root index
  open errors, both routed through the configured 500 error page path.
- Registration contract evidence distinguishes snapshot-capable registrars,
  which get duplicate-route preflight before mutation, from AddRoute-only
  custom registrars, which are explicitly best-effort sequential and can be
  partially registered if a later route add fails.
- Mount inspection edge behavior is explicit: `Mount.Prefix` and
  `Mount.Handler` are nil-receiver safe for inspection, while `Mount.Register`
  rejects nil mounts and registrars.
- Non-`http.Dir` custom filesystem probing is intentionally lazy and may open
  `.br`/`.gz` candidates on original responses to preserve
  `Vary: Accept-Encoding` correctness; directory-backed mounts are the
  recommended path when per-request backend probes are too expensive. Custom
  filesystems can now provide `WithPrecompressedVariantPlan` when they already
  have reliable variant metadata and need to avoid lazy miss probes.
- Directory-backed bundles are treated as immutable deployment artifacts.
  Construction-time variant metadata is deterministic for a mounted release
  directory but is not a runtime atomic snapshot for in-place file mutations.
- Per-request compressed variant failures intentionally remain best-effort
  misses. They downgrade to the original asset when `identity` is acceptable.
  `WithPrecompressedVariantMissHandler` now provides an application-owned signal
  for planned variant open misses and accepted variant stat misses; x/frontend
  still emits no built-in log or metric by default.
- Negotiation parser coverage now exercises shared internal q-value parsing for
  both `Accept` and `Accept-Encoding`.
- Test organization now separates mount, security, compression, response, and
  shared helper coverage.
- Basic benchmarks cover normal asset serving and precompressed asset serving.

## Primer And Boundary State

- Primer: `docs/modules/x-frontend/README.md`
- Manifest: `x/frontend/module.yaml`
- Boundary state: documented and aligned with keeping frontend asset policy in
  `x/frontend`, while stable `router` keeps only primitive static file mounts.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/frontend` API changes.

Release refs:

- none recorded

## API Snapshot Evidence

One current-head baseline snapshot is recorded. It is useful for comparing the
candidate surface during development, but it is not release evidence and does
not clear `api_snapshot_missing` by itself.

The snapshot includes `Option`, but `Option` is a sealed constructor input:
callers should use exported `With*` helpers rather than depending on the
package-private config shape. Future stable-compatible configuration changes
should appear as explicit new helpers and be reviewed in the snapshot diff.

Stable freeze candidates:

- `Registrar`
- `Mount`, including `Prefix`, `Handler`, and `Register`
- `RegisterFromDir` and `RegisterFS`
- `NewMountFromDir`, `NewMountFS`, and `NewHandlerFS`
- sealed `Option`
- `WithPrefix`, `WithIndex`, `WithCacheControl`, `WithIndexCacheControl`,
  `WithFallback`, `WithHeaders`, `WithPrecompressed`, `WithNotFoundPage`,
  `WithErrorPage`, `WithMIMETypes`,
  `WithPrecompressedVariantMissHandler`, `PrecompressedVariantMiss`,
  `WithPrecompressedVariantPlan`, and `PrecompressedVariants`

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/frontend -out /tmp/plumego-x-frontend-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/first-batch/x-frontend-head.snapshot`

## Runtime Contract Decisions

- Precompressed downgrade observability is explicit and application-owned.
  `WithPrecompressedVariantMissHandler` reports planned variant open misses and
  accepted variant stat misses without adding built-in logging, metrics, or
  globals. Nil handler remains the default no-signal behavior.
- Custom filesystem precompressed probing remains lazy by default for backward
  compatibility and `Vary: Accept-Encoding` correctness. Custom integrations
  that already have reliable metadata can provide `WithPrecompressedVariantPlan`
  to avoid lazy `.br`/`.gz` miss probes on original responses.
- Directory-backed bundles remain static deployment artifacts. The
  construction-time variant plan is deterministic for a mounted release
  directory, but x/frontend does not provide a runtime atomic snapshot for
  in-place file mutations.
- AddRoute-only registrars remain best-effort sequential targets with no
  rollback. Snapshot-capable registrars get duplicate-route preflight before
  mutation.
- `Mount.Prefix` and `Mount.Handler` keep nil-receiver inspection behavior.
  `Mount.Register` still rejects nil mounts and nil registrars.

These runtime decisions are implemented and documented, but they still require
frontend owner sign-off before any status promotion.

## Latest Validation

The latest stable-closure pass validated the current head with:

- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`
- `GOCACHE=/private/tmp/plumego-gocache make gates`

The full gate passed on the current head and is useful candidate-state evidence.
It does not replace the missing release history, release-backed API snapshot
comparison, or owner sign-off.

## Release Gate State

Stable promotion still requires a passing repository release gate from the
candidate release state. The current head passes `make gates`, including
boundary checks, vet, race tests, normal tests, stable-root coverage, CLI checks,
and website check/build. Re-run the same gate from the final candidate ref
before any status promotion.

The previously suspected non-frontend `x/mq` risk was rechecked with:

```bash
go test -timeout 20s ./x/mq -run TestKVDeduperLifecycle -count=1
```

The check currently passes, so `TestKVDeduperLifecycle` is not recorded as a
current blocker in this ledger. If a future full release gate fails outside
`x/frontend`, record that exact gate output in the owning module evidence
instead of treating it as an `x/frontend` behavior issue.

## Release Comparison Workflow

Use the release-aware evidence tool when two concrete release refs are
available:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/frontend \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir /tmp/plumego-x-frontend-release-evidence
```

Do not clear `release_history_missing` or `api_snapshot_missing` until the
recorded refs and snapshot files come from real releases.

Current-head snapshots and current-head gates may be refreshed during hardening,
but they must be recorded as development evidence only. A stable promotion
record must name the exact release refs used for comparison and store the
release-backed snapshot output.

## Release Evidence

Not recorded.

Release refs: none recorded

API snapshot comparison: current-head baseline only

## Owner Sign-Off

Missing. The `frontend` owner must confirm the beta criteria before any
`module.yaml` status change.

## Shortest Path To Stable

1. Tag or otherwise identify two concrete consecutive minor release refs that
   include `x/frontend`.
2. Run `extension-release-evidence` between those refs and store the generated
   release-backed API snapshot comparison.
3. Confirm no exported `x/frontend` API churn occurred across those refs.
4. Re-run `GOCACHE=/private/tmp/plumego-gocache make gates` from the final
   candidate ref.
5. Record frontend owner sign-off.
6. Only then change `x/frontend/module.yaml` status.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Posture

Do not promote yet. `x/frontend` remains `experimental`.
