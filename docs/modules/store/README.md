# store

## Purpose

`store` holds persistence primitives and base abstractions.

## Use this module when

- defining stable storage contracts
- implementing basic persistence helpers
- working below topology-heavy data features

## Do not use this module for

- tenant-aware storage policy
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
