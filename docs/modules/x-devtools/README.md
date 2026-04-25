# x/devtools

## Purpose

`x/devtools` provides debug-only routes, profiling endpoints, env reload helpers, and development metrics wiring.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is local debug surface behavior
- the task is pprof, config snapshot, metrics snapshot, or env reload support
- development-only diagnostics need explicit route registration

## Do not use this module for

- production admin policy
- application bootstrap
- default runtime behavior in `core`

## First files to read

- `x/devtools/module.yaml`
- `x/devtools/devtools.go`
- `README.md`

## Main risks when changing this module

- debug routes becoming enabled by default
- hidden runtime side effects
- production behavior accidentally depending on devtools

## Canonical change shape

- keep devtools opt-in
- keep debug handlers explicit and locally mounted
- keep debug runtime snapshot payloads in `x/devtools`, not in `core`
- keep debug JSON success payloads as local typed DTOs instead of one-off maps
- do not move debug routes into `core`
- keep one canonical debug endpoint family under `/_debug/*`

## Canonical debug endpoints

- `/_debug/routes` for plain-text route listing
- `/_debug/routes.json` for machine-readable route discovery
- `/_debug/config` for runtime/config snapshot payloads
- `/_debug/info` for config plus build metadata
- `/_debug/middleware`, `/_debug/metrics`, and `/_debug/reload` for the
  remaining debug-only surfaces

## Boundary with bootstrap

- `x/devtools` is a secondary capability root for local diagnostics, not a bootstrap surface
- keep devtools out of canonical application startup
- mount debug routes explicitly and gate them outside production defaults
