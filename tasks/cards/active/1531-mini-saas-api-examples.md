# Card 1531

Milestone: M-025
Recipe: (none — examples + docs)
Context Package: use-cases/mini-saas-api
Priority: P1
State: active
Primary Module: use-cases/mini-saas-api
Owned Files: api/**, README.md, ARCHITECTURE.md
Depends On: 1528, 1529, 1530

## Goal

The full API is walkable three ways — `api/curl.sh` end to end against a live
server, `api/examples.http` in an editor, and `api/postman_collection.json` in
Postman — and the README/ARCHITECTURE describe the implemented behavior.

## Scope

api/curl.sh, api/examples.http, api/postman_collection.json, README.md
walkthrough + capability table, ARCHITECTURE.md request path + ownership.

## Non-goals

No OpenAPI generation (x/openapi is experimental; possible follow-up card).

## Files

api/curl.sh
api/examples.http
api/postman_collection.json
README.md
ARCHITECTURE.md

## Acceptance Tests

(manual gate) `bash api/curl.sh` against `go run .` prints ALL OK.

## Tests

curl.sh exercises: health, signup, me, tenant usage, idempotent create +
replay, project CRUD, refresh rotation + reuse 401, audit, metrics.

## Docs Sync

README.md, ARCHITECTURE.md (this card).

## Validation

cd use-cases/mini-saas-api && go vet ./... && go test -race -timeout 180s ./...
gofmt -l use-cases/mini-saas-api

## Done Definition

- [x] Acceptance Tests pass (`api/curl.sh` → ALL OK against live server).
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated.

## Outcome

curl walkthrough verified against a live `go run .` instance (port 18090):
all steps pass including idempotency replay match, refresh-reuse 401, and
Prometheus exposition with the mini_saas_api namespace. README gained the
capability table, walkthrough instructions, and auth/idempotency notes;
ARCHITECTURE gained the full request path and new package ownership rows.
