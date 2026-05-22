# with-messaging Scenario Reference

`reference/with-messaging` is a **non-canonical scenario reference**.

It shows how to add `x/messaging` (in-process pub/sub broker) to a service
that follows the same bootstrap structure as `reference/standard-service`.

**This is not the canonical app layout.** See `reference/standard-service` for that.

## What it demonstrates

- Wiring an `x/messaging` in-process broker into the app constructor
- Injecting the broker into a handler via constructor injection
- Registering a publish endpoint alongside the standard health routes
- Keeping the bootstrap shape (`main.run` → `app.Start(ctx)`) aligned with the canonical path
- Loading config with the same precedence as the canonical service:
  `Defaults < .env < process env < flags`

## Design constraints

- depends on the same stable roots as `reference/standard-service`
- also imports `x/messaging` for the broker (intentional — this is a scenario reference)
- keeps `x/messaging` wiring in `internal/app/app.go`, not in `main.go`
- keeps route registration explicit in `internal/app/routes.go`
- keeps process signal ownership in `main.go`; `internal/app` only reacts to the caller-owned context

## Run it

```bash
cd reference/with-messaging
go run .
```

Then publish an event:

```bash
curl -X POST http://localhost:8082/events/publish \
  -H "Content-Type: application/json" \
  -d '{"topic":"user.created","payload":"hello"}'
```
