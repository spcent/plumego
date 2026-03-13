# router

## Purpose

`router` owns route matching, params, groups, and reverse routing.

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

## Canonical change shape

- preserve deterministic dispatch
- keep explicit method-plus-path registration behavior
- avoid bleeding response or middleware policy into router internals
