# with-rest Feature Demo

`reference/with-rest` is a non-canonical feature demo.

It shows how to add `x/rest` resource controllers to a service that still starts
from Plumego's canonical application shape.

`x/rest` remains experimental until beta promotion evidence is complete. Use
this demo as wiring guidance, not as a compatibility claim.

## What It Demonstrates

- normal explicit handlers and `x/rest` resource controllers in the same app
- in-memory repository implementation with no external database
- visible route registration through `core.App`
- standard `contract.WriteResponse` for app-local handlers

## Routes

- `GET /api/hello`
- `GET /api/users`
- `GET /api/users/:id`
- `POST /api/users`

## Run

```bash
go run ./reference/with-rest
```
