# with-rest Scenario Reference

`reference/with-rest` is a non-canonical scenario reference.

It shows how to add `x/rest` resource controllers to a service that still starts
from Plumego's canonical application shape.

`x/rest` remains experimental until beta promotion evidence is complete. Use
this scenario reference as wiring guidance, not as a compatibility claim.

## What It Demonstrates

- normal explicit handlers and `x/rest` resource controllers in the same app
- in-memory repository implementation with no external database
- explicit `x/validate.Bind[T]` request validation with the playground adapter
- visible route registration through `core.App`
- standard `contract.WriteResponse` for app-local handlers

## Routes

- `GET /api/hello`
- `GET /api/users`
- `GET /api/users/:id`
- `POST /api/users`
- `POST /api/items`

## Run

```bash
cd reference/with-rest
go run .
```

Generate the OpenAPI document for the registered routes:

```bash
cd reference/with-rest
make spec
```
