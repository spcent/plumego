# x/webhook

## Purpose

`x/webhook` provides inbound and outbound webhook helpers under the broader messaging family.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is narrow webhook verification or delivery mechanics
- the task is not broader messaging orchestration

## Do not use this module for

- app bootstrap
- general messaging family discovery

## First files to read

- `x/webhook/module.yaml`
- `x/messaging/module.yaml`
- `x/webhook/in.go` or `x/webhook/out.go`

## Public entrypoints

- `NewService`
- `NewMemStore`
- `VerifyHMAC`
- `VerifyGitHub`
- `VerifyStripe`

## Main risks when changing this module

- signature verification regression
- delivery retry regression
- route registration regression

## Canonical change shape

- keep verification fail-closed
- keep secrets and signatures out of logs
- keep webhook transport concerns explicit and local to this layer
