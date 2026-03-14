# store

## Purpose

`store` holds persistence primitives and base abstractions.

## Use this module when

- defining stable storage contracts
- implementing basic persistence helpers
- working below topology-heavy data features

## Do not use this module for

- tenant-aware storage policy
- tenant-aware adapters
- sharding or heavy topology defaults in stable roots
- app bootstrap

## First files to read

- `store/module.yaml`
- the target package under `store/*`
- `specs/repo.yaml`

## Canonical change shape

- keep interfaces narrow
- keep concurrent behavior testable
- move topology-heavy features to owning extensions

## Current migration debt

The stable `store` root still contains two packages that should be treated as migration debt rather than a pattern to extend:

- `store/cache/distributed`
- `store/cache/redis`

Current rule:

- keep these packages working
- do not add new topology-heavy or provider-heavy siblings under stable `store`
- do not add tenant-aware adapters or tenant-specific storage policy under stable `store`
- route new topology-heavy data capabilities to `x/data` or `x/tenant`
