# core

## Purpose

`core` is the HTTP application kernel. It owns app construction, route attachment, middleware attachment, and server lifecycle.

## Use this module when

- assembling an application from explicit routes and middleware
- starting or stopping an HTTP server
- wiring shared runtime facilities such as logging, health, and metrics

## Do not use this module for

- route matching
- feature catalogs or plugin containers
- tenant policy
- persistence behavior

## First files to read

- `core/module.yaml`
- `core/app.go`
- `core/options.go`
- `reference/standard-service/internal/app/app.go`

## Canonical change shape

- keep bootstrap explicit
- keep lifecycle behavior reviewable
- push feature-specific wiring back to app-local code or the owning extension
