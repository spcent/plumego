# Config Module

> **Status**: Legacy public root removed

## Current Rule

Plumego no longer exposes a public `config/` root package.

Configuration loading is application-owned:

- keep env/file/flag parsing in your `main` package or app-local wiring
- pass concrete values into `core.New(...)` and other constructors
- keep shared helpers inside app-local `internal/config`

## Canonical Locations

- reference applications: `reference/.../internal/config`
- framework internals: `internal/config`
- user-facing runtime options: `core.With...` plus explicit env/flag parsing in your app

## Migration Guidance

If you previously imported the removed public config root, migrate to one of these patterns:

1. Parse environment variables directly in `main`.
2. Define an `AppConfig` struct owned by your application.
3. Keep reusable config loaders inside your own `internal/config`.

## Scope Note

The documents under this directory are legacy migration notes. They are not authoritative for new application imports.
