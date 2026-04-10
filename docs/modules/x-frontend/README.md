# x/frontend

## Purpose

`x/frontend` provides explicit helpers for static asset serving, embedded frontend mounts, and SPA fallback behavior.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

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
- fallback regression
- `net/http` incompatibility in mounting behavior

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
