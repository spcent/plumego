# x/websocket

## Purpose

`x/websocket` owns websocket server helpers and explicit route registration for websocket transport.

## Use this module when

- the task is websocket transport behavior
- the change is connection lifecycle or hub behavior

## Do not use this module for

- app bootstrap
- generic HTTP routing outside websocket transport

## First files to read

- `x/websocket/module.yaml`
- `x/websocket/websocket.go`
- `reference/standard-service` when checking canonical bootstrap shape
