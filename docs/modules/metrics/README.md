# metrics

## Purpose

`metrics` holds metrics interfaces and base collectors.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- changing collector contracts
- adding base metrics helpers
- instrumenting stable library behavior

## Do not use this module for

- exporter brand ownership
- gateway policy
- app bootstrap

## First files to read

- `metrics/module.yaml`
- `metrics/*.go`
- owning extension docs when the change is exporter-specific

## Canonical change shape

- keep collector APIs small
- keep exporters and integration wiring in owning extensions
