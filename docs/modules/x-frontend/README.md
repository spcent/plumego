# x/frontend

## Purpose

`x/frontend` provides explicit helpers for static asset serving, embedded frontend mounts, and SPA fallback behavior.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen
- Hardened coverage exists for path safety, symlink escape rejection,
  precompressed negotiation, cache variance, custom page policy, navigation-only
  fallback, and basic mount behavior. Basic benchmarks cover normal and
  precompressed asset serving.
- Current-head API snapshot evidence exists, but stable promotion still requires
  release-backed snapshot comparisons, release history, and owner sign-off

## Use this module when

- the task is static asset serving
- the task is embedded frontend mounting
- the task is fallback behavior for single-page applications
- the task needs frontend asset policy such as cache headers, custom headers, precompressed assets, or MIME overrides

## Do not use this module for

- application bootstrap
- business-specific UI flow logic
- hidden filesystem registration in `core`
- route matching primitives; keep those in `router`

## First files to read

- `x/frontend/module.yaml`
- `x/frontend/README.md`
- `specs/task-routing.yaml`

## Main risks when changing this module

- hidden filesystem side effects
- process working-directory changes affecting directory-backed mounts
- fallback regression
- missing asset requests accidentally returning SPA index responses
- directory symlink escape regression
- `net/http` incompatibility in mounting behavior
- precompressed response cache variance regression

## Canonical change shape

- keep frontend mounting explicit
- preserve transport-only helper behavior
- do not let frontend helpers redefine the canonical app path
- own frontend asset policy here instead of adding policy knobs to stable `router`

## Boundary with bootstrap

- `x/frontend` is a secondary capability root for asset serving, not an application bootstrap surface
- read `reference/standard-service` only to align app shape and route wiring
- keep frontend registration explicit in app-local code

## Boundary with router

- `router.Static` and `router.StaticFS` are small stable file-mount primitives.
- `x/frontend` owns higher-level frontend serving behavior: SPA fallback, cache headers, custom headers, precompressed files, custom error pages, and MIME overrides.
- Do not add frontend asset policy knobs back to stable `router`; add or refine them here.

## Registration Contract

- Root mounts register ANY `/` and ANY `/*filepath`.
- Prefixed mounts register ANY `<prefix>/*filepath` and ANY `<prefix>`.
- Registrars that expose route snapshots are preflighted for duplicate target
  routes before mutation.
- AddRoute-only custom registrars remain best-effort sequential targets.

## Header Policy

- Use `WithHeaders` for security and metadata headers only.
- Use dedicated options for cache, MIME, and precompressed response semantics.
- Do not allow caller-provided headers to override internally managed
  `Content-*`, `Vary`, cache, conditional, range, or hop-by-hop response
  semantics.

## Encoding Negotiation

- `x/frontend` serves precompressed `.br` and `.gz` variants only when requested
  and available.
- Directory-backed mounts, including `http.Dir` inputs passed to `RegisterFS`,
  build immutable precompressed variant metadata at construction time and fail
  mount construction when the variant scan hits filesystem errors.
- Non-`http.Dir` caller-provided filesystems keep lazy variant probing.
- Invalid `Accept-Encoding` quality values invalidate that token rather than
  being clamped.
- Requests that refuse `identity` receive `406 Not Acceptable` when no accepted
  precompressed variant can be served.

## Filesystem Contract

- Directory mounts and `http.Dir` inputs resolve the configured root to an
  absolute canonical path at construction time.
- Directory mounts and `http.Dir` inputs fail fast when the configured index
  file is missing or is a directory.
- Non-`http.Dir` caller-provided `http.FileSystem` implementations remain lazy
  and own their backend readiness and storage boundaries.

## Stable-readiness blockers

- `status` remains `experimental`; do not promote without owner approval.
- The option API is intentionally sealed: callers compose exported `With*`
  options, and new stable knobs should be added as explicit helpers rather than
  caller-defined custom options against internal config state.
- Public API compatibility still needs a release snapshot comparison for
  `Registrar`, `Mount`, `Option`, `NewMountFromDir`, `NewMountFS`,
  `NewHandlerFS`, `RegisterFromDir`, `RegisterFS`, and the `With*` options.
- Current-head snapshot:
  `docs/extension-evidence/snapshots/first-batch/x-frontend-head.snapshot`.
- Release evidence must show at least two consecutive minor releases without
  exported API churn.
- Embedded helper guidance must stay explicit: applications pass their own
  `embed.FS` through `RegisterFS`.
