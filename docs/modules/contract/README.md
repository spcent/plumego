# contract

## Purpose

`contract` defines structured transport contracts: request context helpers, API errors, and response helpers.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- standardizing error response shape
- writing transport-level response helpers
- carrying request metadata needed by handlers

## Do not use this module for

- protocol gateway families
- business DTO ownership
- route matching

## First files to read

- `contract/module.yaml`
- `contract/error.go`
- `contract/response.go`

## Canonical change shape

- preserve one clear error-writing path
- keep helpers transport-focused
- avoid framework-style abstraction layers
