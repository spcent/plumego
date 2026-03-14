# router

## Purpose

`router` owns route matching, params, groups, and reverse routing.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- adding matching or grouping behavior
- changing path parameter extraction
- working on reverse routing or mount primitives

## Do not use this module for

- JSON responses
- auth policy
- service construction

## First files to read

- `router/module.yaml`
- `router/router.go`
- `router/group.go`

## Public entrypoints

- `Router`
- `Group`
- `Param`

## Main risks when changing this module

- dispatch regression
- param extraction regression
- reverse routing regression

## Canonical change shape

- preserve deterministic dispatch
- keep explicit method-plus-path registration behavior
- avoid bleeding response or middleware policy into router internals
