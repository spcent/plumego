# x/fileapi

## Purpose

`x/fileapi` is the app-facing transport surface for tenant-aware file upload, download, metadata lookup, and temporary URL endpoints.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is file upload or download HTTP behavior
- the task is multipart parsing, file response headers, or file URL transport
- the task is request-context-based tenant extraction for file APIs

## Do not use this module for

- pure storage interfaces
- tenant-aware storage backend implementation
- application bootstrap
- business-specific file workflows

## First files to read

- `x/fileapi/module.yaml`
- `x/fileapi/handler.go`
- `specs/task-routing.yaml`

## Main risks when changing this module

- tenant extraction regression
- multipart request parsing regression
- streaming response regression
- non-canonical error handling

## Canonical change shape

- keep handler behavior transport-only
- read tenant identity from request context, not body or path
- delegate storage and metadata behavior to `x/data/file`
- keep stable storage contracts, shared types, and errors in `store/file`
- keep route wiring in the application layer; `x/fileapi` should not become a bootstrap surface

## Boundary with data and store

- `x/fileapi` is the app-facing file transport entrypoint
- `x/data/file` owns tenant-aware storage, metadata, and temporary URL implementations
- `store/file` owns transport-agnostic interfaces, errors, and shared types
- if the change is about multipart parsing, status codes, or response headers, keep it here
- if the change is about tenant path layout, metadata persistence, or storage backends, move it to `x/data/file`
- if the change is about stable storage contracts or shared file types, keep it in `store/file`
