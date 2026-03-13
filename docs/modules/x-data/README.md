# x/data

## Purpose

`x/data` contains topology-heavy and fast-evolving data capabilities such as sharding and advanced rw patterns.

## Use this module when

- the task is beyond stable `store` primitives
- sharding or topology-aware behavior is involved

## Do not use this module for

- stable store contracts
- application bootstrap

## First files to read

- `x/data/module.yaml`
- the owning subpackage under `x/data/*`
- `specs/repo.yaml`
