# Request Decode Example

This example demonstrates canonical explicit JSON decode + validation in handlers.

## Run

```bash
go run ./examples/bind-example
```

The service listens on `:8082`.

## Success

```bash
curl -X POST http://127.0.0.1:8082/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"supersecret"}'
```

Expected response:

```json
{"email":"user@example.com","ok":true}
```

## Validation error (field-level)

```bash
curl -X POST http://127.0.0.1:8082/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"not-an-email","password":""}'
```

Expected response (example):

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "category": "validation_error",
    "details": {
      "fields": [
        {"field":"email","code":"email","message":"invalid email format"},
        {"field":"password","code":"required","message":"is required"}
      ]
    }
  }
}
```

## JSON parse error / unknown fields

```bash
curl -X POST http://127.0.0.1:8082/v1/users \
  -H 'Content-Type: application/json' \
  -d '{'
```

```bash
curl -X POST http://127.0.0.1:8082/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"secret","extra":"x"}'
```

Expected response code: `INVALID_JSON`.

## Style rule

Do not inject business DTOs via middleware and do not retrieve DTOs from request context in new code.
Use explicit `json.NewDecoder(r.Body).Decode(&req)` plus local validation directly in handlers.
